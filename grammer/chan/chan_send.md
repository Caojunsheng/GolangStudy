
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
eyJoaXN0b3J5IjpbLTE5NjY0NDk2NDFdfQ==
-->