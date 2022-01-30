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

## 二、chan源码解读
### 1、chan数据结构
```go
type hchan struct {
    // chan 里元素数量
    qcount   uint
    // chan 底层循环数组的长度
    dataqsiz uint
    // 指向底层循环数组的指针
    // 只针对有缓冲的 channel
    buf      unsafe.Pointer
    // chan 中元素大小
    elemsize uint16
    // chan 是否被关闭的标志
    closed   uint32
    // chan 中元素类型
    elemtype *_type // element type
    // 已发送元素在循环数组中的索引
    sendx    uint   // send index
    // 已接收元素在循环数组中的索引
    recvx    uint   // receive index
    // 等待接收的 goroutine 队列
    recvq    waitq  // list of recv waiters
    // 等待发送的 goroutine 队列
    sendq    waitq  // list of send waiters
    // 保护 hchan 中所有字段
    lock mutex
}
type waitq struct {  
   first *sudog  
   last *sudog  
}
```
`buf`  指向底层循环数组，只有缓冲型的 channel 才有。

`sendx`，`recvx`  均指向底层循环数组，表示当前可以发送和接收的元素位置索引值（相对于底层数组）。

`sendq`，`recvq`  分别表示被阻塞的 goroutine，这些 goroutine 由于尝试读取 channel 或向 channel 发送数据而被阻塞。

`waitq`  是  `sudog`  的一个双向链表，而  `sudog`  实际上是对 goroutine 的一个封装。

例如，创建一个容量为 6 的，元素为 int 型的 channel 数据结构如下 ：
![enter image description here](https://static.sitestack.cn/projects/qcrao-Go-Questions/47e89d2a3dd43e867b808a10576c8271.png)

### 2、chan读取源码
编译器处理完之后，chan的读取在go中入口是下面两个函数：
```go
// 读取的数据放在elem里面，两种读取的方式，第一种直接返回值，第二种返回一个bool值，判断chan是否关闭
func chanrecv1(c *hchan, elem unsafe.Pointer) {
    chanrecv(c, elem, true)
}
func chanrecv2(c *hchan, elem unsafe.Pointer) (received bool) {
    _, received = chanrecv(c, elem, true)
    return
}
```
`chanrecv1`不返回ok，`chanrecv2`返回ok，两个最终都是调用`chanrecv`函数
```go
// src/runtime/chan.go:454
// chanrecv receives on channel c and writes the received data to ep.// ep may be nil, in which case received data is ignored.  
// If block == false and no elements are available, returns (false, false).// Otherwise, if c is closed, zeros *ep and returns (true, false).  
// Otherwise, fills in *ep with an element and returns (true, true).  
// A non-nil ep must point to the heap or the caller's stack.  
func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {  
   // raceenabled: don't need to check ep, as it is always on the stack  
 // or is new memory allocated by reflect.  
  if debugChan {  
      print("chanrecv: chan=", c, "\n")  
   }  
  
  // 如果chan是nil的话
   if c == nil {
      // 非阻塞调用，则直接返回false, false  
      if !block {  
         return  
  }  
      // 阻塞调用，一直等待接收nil的chan，goroutine挂起
      gopark(nil, nil, waitReasonChanReceiveNilChan, traceEvGoStop, 2)  
      throw("unreachable")  
   }  
  
   // Fast path: check for failed non-blocking operation without acquiring the lock.  
  if !block && empty(c) {  
      // After observing that the channel is not ready for receiving, we observe whether the  
 // channel is closed. // // Reordering of these checks could lead to incorrect behavior when racing with a close. // For example, if the channel was open and not empty, was closed, and then drained, // reordered reads could incorrectly indicate "open and empty". To prevent reordering, // we use atomic loads for both checks, and rely on emptying and closing to happen in // separate critical sections under the same lock.  This assumption fails when closing // an unbuffered channel with a blocked send, but that is an error condition anyway.  if atomic.Load(&c.closed) == 0 {  
         // Because a channel cannot be reopened, the later observation of the channel  
 // being not closed implies that it was also not closed at the moment of the // first observation. We behave as if we observed the channel at that moment // and report that the receive cannot proceed.  return  
  }  
      // The channel is irreversibly closed. Re-check whether the channel has any pending data  
 // to receive, which could have arrived between the empty and closed checks above. // Sequential consistency is also required here, when racing with such a send.  if empty(c) {  
         // The channel is irreversibly closed and empty.  
  if raceenabled {  
            raceacquire(c.raceaddr())  
         }  
         if ep != nil {  
            typedmemclr(c.elemtype, ep)  
         }  
         return true, false  
  }  
   }  
  
   var t0 int64  
 if blockprofilerate > 0 {  
      t0 = cputicks()  
   }  
  
   lock(&c.lock)  
  
   if c.closed != 0 && c.qcount == 0 {  
      if raceenabled {  
         raceacquire(c.raceaddr())  
      }  
      unlock(&c.lock)  
      if ep != nil {  
         typedmemclr(c.elemtype, ep)  
      }  
      return true, false  
  }  
  
   if sg := c.sendq.dequeue(); sg != nil {  
      // Found a waiting sender. If buffer is size 0, receive value  
 // directly from sender. Otherwise, receive from head of queue // and add sender's value to the tail of the queue (both map to // the same buffer slot because the queue is full).  recv(c, sg, ep, func() { unlock(&c.lock) }, 3)  
      return true, true  
  }  
  
   if c.qcount > 0 {  
      // Receive directly from queue  
  qp := chanbuf(c, c.recvx)  
      if raceenabled {  
         racenotify(c, c.recvx, nil)  
      }  
      if ep != nil {  
         typedmemmove(c.elemtype, ep, qp)  
      }  
      typedmemclr(c.elemtype, qp)  
      c.recvx++  
      if c.recvx == c.dataqsiz {  
         c.recvx = 0  
  }  
      c.qcount--  
      unlock(&c.lock)  
      return true, true  
  }  
  
   if !block {  
      unlock(&c.lock)  
      return false, false  
  }  
  
   // no sender available: block on this channel.  
  gp := getg()  
   mysg := acquireSudog()  
   mysg.releasetime = 0  
  if t0 != 0 {  
      mysg.releasetime = -1  
  }  
   // No stack splits between assigning elem and enqueuing mysg  
 // on gp.waiting where copystack can find it.  mysg.elem = ep  
  mysg.waitlink = nil  
  gp.waiting = mysg  
 mysg.g = gp  
 mysg.isSelect = false  
  mysg.c = c  
  gp.param = nil  
  c.recvq.enqueue(mysg)  
   // Signal to anyone trying to shrink our stack that we're about  
 // to park on a channel. The window between when this G's status // changes and when we set gp.activeStackChans is not safe for // stack shrinking.  atomic.Store8(&gp.parkingOnChan, 1)  
   gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanReceive, traceEvGoBlockRecv, 2)  
  
   // someone woke us up  
  if mysg != gp.waiting {  
      throw("G waiting list is corrupted")  
   }  
   gp.waiting = nil  
  gp.activeStackChans = false  
 if mysg.releasetime > 0 {  
      blockevent(mysg.releasetime-t0, 2)  
   }  
   success := mysg.success  
  gp.param = nil  
  mysg.c = nil  
  releaseSudog(mysg)  
   return true, success  
}
```
<!--stackedit_data:
eyJoaXN0b3J5IjpbLTEzMjg3Mjc3MDEsLTE4MjA0NDU3MF19
-->