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





### 4、chan关闭流程
<!--stackedit_data:
eyJoaXN0b3J5IjpbNjg2Nzk1OTk5XX0=
-->