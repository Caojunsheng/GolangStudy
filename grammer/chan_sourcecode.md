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

### 2、chan读取源码分析
chan的读取源码入口是如下两个函数：
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
// chanrecv 函数接收 channel c 的元素并将其写入 ep 所指向的内存地址。
// 如果 ep 是 nil，说明忽略了接收值。
// 如果 block == false，即非阻塞型接收，在没有数据可接收的情况下，返回 (false, false)
// 否则，如果 c 处于关闭状态，将 ep 指向的地址清零，返回 (true, false)
// 否则，用返回值填充 ep 指向的内存地址。返回 (true, true)
// 如果 ep 非空，则应该指向堆或者函数调用者的栈
func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
	...
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

	// 如果是非阻塞且chan是空的
	if !block && empty(c) {
	    // 如果chan是未关闭的，直接返回false,false
		if atomic.Load(&c.closed) == 0 {
			return
		}
		// chan已经关闭，并且为空，老实说。这段代码感觉有点多余，下面也处理了这种情况
		if empty(c) {
			// The channel is irreversibly closed and empty.
			if raceenabled {
				raceacquire(c.raceaddr())
			}
			if ep != nil {
			    // 对于已关闭的chan执行接收，不忽略返回值的情况下，会受到该类型的零值，清理ep的内存
				typedmemclr(c.elemtype, ep)
			}
			// 返回selected为true
			return true, false
		}
	}

	var t0 int64
	if blockprofilerate > 0 {
		t0 = cputicks()
	}

	lock(&c.lock)
    // chan已经关闭，且缓存中无数据，直接返回该类型的零值
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
		// 如果sender中有等待发送，那么可以分为两种情况
		// 1、非缓冲队列，即同步chan，则直接从sender中接收值。
		// 2、缓冲队列，即异步chan，从缓冲队列的头部拷贝到接收者，拷贝发送队列的值到缓冲队列末尾
		recv(c, sg, ep, func() { unlock(&c.lock) }, 3)
		return true, true
	}

    // 缓冲型chan，buf里面有元素，直接从buf里面拿
	if c.qcount > 0 {
		// Receive directly from queue
		qp := chanbuf(c, c.recvx)
		if raceenabled {
			racenotify(c, c.recvx, nil)
		}
		// 代码里面需要接收值，则需要拷贝值，比如接收是`val<-ch`，而不是`<-ch`，需要把chan的值拷贝到val
		if ep != nil {
			typedmemmove(c.elemtype, ep, qp)
		}
		// 清空掉原来buf中对应位置的值
		typedmemclr(c.elemtype, qp)
		// 接收index+1
		c.recvx++
		// 如果接收索引已经到末尾，重新移到队首
		if c.recvx == c.dataqsiz {
			c.recvx = 0
		}
		// 缓冲区大小减一
		c.qcount--
		// 解锁
		unlock(&c.lock)
		return true, true
	}

	if !block {
	    // 非阻塞接收，解锁，返回false,false
		unlock(&c.lock)
		return false, false
	}

	// 无发送者，这个接收值需要被阻塞.
	gp := getg()
	mysg := acquireSudog()
	mysg.releasetime = 0
	if t0 != 0 {
		mysg.releasetime = -1
	}
	// 构造一个接收数据的sudog.
	mysg.elem = ep
	mysg.waitlink = nil
	gp.waiting = mysg
	mysg.g = gp
	mysg.isSelect = false
	mysg.c = c
	gp.param = nil
	// 放入接受者队列中
	c.recvq.enqueue(mysg)
	// 将goroutine挂起
	atomic.Store8(&gp.parkingOnChan, 1)
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
empty源码分析
 1. 如果是非缓冲型，且sendq中无goroutine

 2. 缓冲型，但是buf里面没有元素

```go
func empty(c *hchan) bool {
	// c.dataqsiz is immutable.
	if c.dataqsiz == 0 {
		return atomic.Loadp(unsafe.Pointer(&c.sendq.first)) == nil
	}
	return atomic.Loaduint(&c.qcount) == 0
}
```

### 3、chan写入源码分析

```go
func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
    // 如果chan是空
	if c == nil {
	    // 非阻塞，直接返回false，表示未发送成功
		if !block {
			return false
		}
		// 阻塞的，挂起goroutine
		gopark(nil, nil, waitReasonChanSendNilChan, traceEvGoStop, 2)
		throw("unreachable")
	}

	...
	// 如果是非阻塞的，chan未关闭，且chan的buffer已经满了，则返回发送失败
	if !block && c.closed == 0 && full(c) {
		return false
	}

	var t0 int64
	if blockprofilerate > 0 {
		t0 = cputicks()
	}

	lock(&c.lock)

	if c.closed != 0 {
		unlock(&c.lock)
		// 如果chan已经关闭了，再向chan发送数据，直接报panic
		panic(plainError("send on closed channel"))
	}

	if sg := c.recvq.dequeue(); sg != nil {
		// 如果有接受者在等待，直接将发送的数据拷贝到
		send(c, sg, ep, func() { unlock(&c.lock) }, 3)
		return true
	}

    // 如果缓冲的chan，还有空间，将发送的数据拷贝到buffer中
	if c.qcount < c.dataqsiz {
		qp := chanbuf(c, c.sendx)
		if raceenabled {
			racenotify(c, c.sendx, nil)
		}
		typedmemmove(c.elemtype, qp, ep)
		// 发送游标+1
		c.sendx++
		// 发送游标已经到末尾了，重新移到队头
		if c.sendx == c.dataqsiz {
			c.sendx = 0
		}
		// 缓冲区数量+1
		c.qcount++
		unlock(&c.lock)
		return true
	}
    // 非阻塞的chan，直接返回写入失败
	if !block {
		unlock(&c.lock)
		return false
	}

	// chan满了，发送者会被阻塞，构造一个sudog挂起
	gp := getg()
	mysg := acquireSudog()
	mysg.releasetime = 0
	if t0 != 0 {
		mysg.releasetime = -1
	}

	mysg.elem = ep
	mysg.waitlink = nil
	mysg.g = gp
	mysg.isSelect = false
	mysg.c = c
	gp.waiting = mysg
	gp.param = nil
	// 构造sudog放入发送者队列
	c.sendq.enqueue(mysg)
	atomic.Store8(&gp.parkingOnChan, 1)
	gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanSend, traceEvGoBlockSend, 2)
	KeepAlive(ep)

	// someone woke us up.
	if mysg != gp.waiting {
		throw("G waiting list is corrupted")
	}
	gp.waiting = nil
	gp.activeStackChans = false
	closed := !mysg.success
	gp.param = nil
	if mysg.releasetime > 0 {
		blockevent(mysg.releasetime-t0, 2)
	}
	mysg.c = nil
	releaseSudog(mysg)
	if closed {
		if c.closed == 0 {
			throw("chansend: spurious wakeup")
		}
		panic(plainError("send on closed channel"))
	}
	return true
}
```
send源码
```go
// send 函数处理向一个空的 channel 发送操作
// ep 指向被发送的元素，会被直接拷贝到接收的 goroutine
// 之后，接收的 goroutine 会被唤醒
// c 必须是空的（因为等待队列里有 goroutine，肯定是空的）
// c 必须被上锁，发送操作执行完后，会使用 unlockf 函数解锁
// sg 必须已经从等待队列里取出来了
// ep 必须是非空，并且它指向堆或调用者的栈
func send(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int) {
    // 省略一些用不到的
    // ……
    // sg.elem 指向接收到的值存放的位置，如 val <- ch，指的就是 &val
    if sg.elem != nil {
        // 直接拷贝内存（从发送者到接收者）
        sendDirect(c.elemtype, sg, ep)
        sg.elem = nil
    }
    // sudog 上绑定的 goroutine
    gp := sg.g
    // 解锁
    unlockf()
    gp.param = unsafe.Pointer(sg)
    if sg.releasetime != 0 {
        sg.releasetime = cputicks()
    }
    // 唤醒接收的 goroutine. skip 和打印栈相关，暂时不理会
    goready(gp, skip+1)
}
```
full源码
```go
func full(c *hchan) bool {
	// 非缓冲chan，判断recvq为空，则认为满
	if c.dataqsiz == 0 {
		return c.recvq.first == nil
	}
	// 缓冲chan，缓冲区数量等于chan大小
	return c.qcount == c.dataqsiz
}
```


<!--stackedit_data:
eyJoaXN0b3J5IjpbMTUyNTE4MDYsMTAxNjU1ODA3NSwtMzQ0NT
Y1NjAzLDEyMzU3MDcyMDZdfQ==
-->