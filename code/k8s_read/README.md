### 1、kubernetes panic堆栈打印
```
func HandlePanic(fn func()) func() {  
   return func() {  
      defer func() {  
         if r := recover(); r != nil {  
            for _, fn := range utilruntime.PanicHandlers {  
               fn(r)  
            }  
            panic(r)  
         }  
      }()  
      // call the function  
  fn()  
   }  
}
```

```go
func logPanic(r interface{}) {  
   if r == http.ErrAbortHandler {  
      // honor the http.ErrAbortHandler sentinel panic value:  
 //   ErrAbortHandler is a sentinel panic value to abort a handler. //   While any panic from ServeHTTP aborts the response to the client, //   panicking with ErrAbortHandler also suppresses logging of a stack trace to the server's error log.  return  
  }  
  
   // Same as stdlib http server code. Manually allocate stack trace buffer size  
 // to prevent excessively large logs  const size = 64 << 10  
  stacktrace := make([]byte, size)  
   stacktrace = stacktrace[:runtime.Stack(stacktrace, false)]  
   if _, ok := r.(string); ok {  
      klog.Errorf("Observed a panic: %s\n%s", r, stacktrace)  
   } else {  
      klog.Errorf("Observed a panic: %#v (%v)\n%s", r, r, stacktrace)  
   }  
}
```


> Written with [StackEdit](https://stackedit.io/).
<!--stackedit_data:
eyJoaXN0b3J5IjpbLTE4NTgyODA5NTIsLTE1MTAxMTU3MSw3Mz
A5OTgxMTZdfQ==
-->