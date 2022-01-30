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

编译器处理完之后，chan的读取在go中入口是下面两个函数：
```go
// 读取的数据放在elem里面，两种读取的方式，diyizhong
func chanrecv1(c *hchan, elem unsafe.Pointer) {
    chanrecv(c, elem, true)
}
func chanrecv2(c *hchan, elem unsafe.Pointer) (received bool) {
    _, received = chanrecv(c, elem, true)
    return
}
```
```go
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
  
   if c == nil {  
      if !block {  
         return  
  }  
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
eyJoaXN0b3J5IjpbLTE4MDMyNzIzOCwzNzk1MzM1OCwtMTYzMj
EzMzczMF19
-->