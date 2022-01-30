### 1、向channel发送数据

向已关闭的channnel发送数据会panic

```go
// src/runtime/chan.go:202
if c.closed != 0 {  
   unlock(&c.lock)  
   panic(plainError("send on closed channel"))  
}
```
### 2、关闭channel

-   关闭nil的channel，会panic
-   对已关闭的channel，再次关闭channel，会panic

> Blockquote
> 
> 
> 
> 
> 
> 
> 
> 
> Written with [StackEdit](https://stackedit.io/).

<!--stackedit_data:
eyJoaXN0b3J5IjpbLTI3OTYxNjg5M119
-->