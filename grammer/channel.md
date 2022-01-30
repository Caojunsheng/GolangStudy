# 1、向channel发送数据

向已关闭的channnel发送数据会panic

```go
// src/runtime/chan.go:202
if c.closed != 0 {  
   unlock(&c.lock)  
   panic(plainError("send on closed channel"))  
}
```

> Written with [StackEdit](https://stackedit.io/).
<!--stackedit_data:
eyJoaXN0b3J5IjpbLTE5MDcwOTE5MjRdfQ==
-->