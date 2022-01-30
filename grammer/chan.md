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



<!--stackedit_data:
eyJoaXN0b3J5IjpbMTMxMjkyNTQ4NywxODE1Njc1ODcwLC0xOD
IwNDQ1NzBdfQ==
-->