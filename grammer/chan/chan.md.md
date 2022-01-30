## 一、golang里面几个chan常见的坑

### 1、向chan发送数据

向已关闭的chan发送数据会panic

```go
// src/runtime/chan.go:202
if c.closed != 0 {  
   unlock(&c.lock)  
   panic(plainError("send on closed channel"))  
}
```
### 2、关闭chan

-   关闭nil的chan，会panic
-   对已关闭的chan，再次关闭chan，会panic

```go
// src/runtime/chan.go:355
func closechan(c *hchan) {  
   if c == nil {  
      panic(plainError("close of nil channel"))  
   }  
  
   lock(&c.lock)  
   if c.closed != 0 {  
      unlock(&c.lock)  
      panic(plainError("close of closed channel"))  
   }
   ...
}
```
### 3、读chan数据

-   chan关闭之后，关闭前放入的数据，仍然可以读取
-   已关闭的chan仍然可以读取，值为零值，返回值ok为false


**如何优雅的关闭channel？**

根据 sender 和 receiver 的个数，分下面几种情况：

1.  一个 sender，一个 receiver
2.  一个 sender， M 个 receiver
3.  N 个 sender，一个 reciver
4.  N 个 sender， M 个 receiver

1,2两种情况，仅一个sender，直接从sender端关闭channel即可

第3种情况解决方案就是增加一个传递关闭信号的 channel，receiver 通过信号 channel 下达关闭数据 channel 指令。senders 监听到关闭信号后，停止发送数据
```go
func main() {
    rand.Seed(time.Now().UnixNano())
    const Max = 100000
    const NumSenders = 1000
    dataCh := make(chan int, 100)
    stopCh := make(chan struct{})
    // senders
    for i := 0; i < NumSenders; i++ {
        go func() {
            for {
                select {
                case <- stopCh:
                    return
                case dataCh <- rand.Intn(Max):
                }
            }
        }()
    }
    // the receiver
    go func() {
        for value := range dataCh {
            if value == Max-1 {
                fmt.Println("send stop signal to senders.")
                close(stopCh)
                return
            }
            fmt.Println(value)
        }
    }()
    select {
    case <- time.After(time.Hour):
    }
}
```
对于第4种情况，和第 3 种情况不同，这里有 M 个 receiver，如果直接还是采取第 3 种解决方案，由 receiver 直接关闭 stopCh 的话，就会重复关闭一个 channel，导致 panic。因此需要增加一个中间人，M 个 receiver 都向它发送关闭 dataCh 的“请求”，中间人收到第一个请求后，就会直接下达关闭 dataCh 的指令（通过关闭 stopCh，这时就不会发生重复关闭的情况，因为 stopCh 的发送方只有中间人一个）。另外，这里的 N 个 sender 也可以向中间人发送关闭 dataCh 的请求。
```go
func main() {
    rand.Seed(time.Now().UnixNano())
    const Max = 100000
    const NumReceivers = 10
    const NumSenders = 1000
    dataCh := make(chan int, 100)
    stopCh := make(chan struct{})
    // It must be a buffered channel.
    toStop := make(chan string, 1)
    var stoppedBy string
    // moderator
    go func() {
        stoppedBy = <-toStop
        close(stopCh)
    }()
    // senders
    for i := 0; i < NumSenders; i++ {
        go func(id string) {
            for {
                value := rand.Intn(Max)
                if value == 0 {
                    select {
                    case toStop <- "sender#" + id:
                    default:
                    }
                    return
                }
                select {
                case <- stopCh:
                    return
                case dataCh <- value:
                }
            }
        }(strconv.Itoa(i))
    }
    // receivers
    for i := 0; i < NumReceivers; i++ {
        go func(id string) {
            for {
                select {
                case <- stopCh:
                    return
                case value := <-dataCh:
                    if value == Max-1 {
                        select {
                        case toStop <- "receiver#" + id:
                        default:
                        }
                        return
                    }
                    fmt.Println(value)
                }
            }
        }(strconv.Itoa(i))
    }
    select {
    case <- time.After(time.Hour):
    }
}
```
<!--stackedit_data:
eyJoaXN0b3J5IjpbMTAxNTI3OTc0M119
-->