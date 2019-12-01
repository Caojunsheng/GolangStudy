### 1. golang pointer address
当需要修改一个指针的值，又无法直接给它赋值，通过函数去修改的时候，可以通过下面方式修改变量的值.

refer to [gopointer.go](https://github.com/Caojunsheng/GolangStudy/blob/master/code/basics/gopointer.go)
```
package main

import "fmt"

var str []string

func main() {
	setVal(&str)
	fmt.Println(str)

	var testStr string
	changeStr(&testStr)
	fmt.Println(testStr)
}

//需要在这里赋值str，但是又不能直接引用 str
func setVal(val *[]string) {
	*val = []string{"a", "b"}
}

func changeStr(str *string) {
	strTemp := "hhh"
	*str = strTemp
}

```
### 2. golang select specific character
golang语言的select特殊特性，当两个select的case同时满足的时候，golang只会选择其中的一个执行，另一个无法执行到。
如果你有两个定时任务，同时触发，那么这时候只会有其中一个被触发。

如果没有任何一个case满足，那么将会执行default的语句。


refer to [select.go](./select.go)

```
package main

import (
	"fmt"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(1)
	int_chan := make(chan int, 1)
	string_chan := make(chan string, 1)
	int_chan <- 1
	string_chan <- "hello"
	select {
	case value := <-int_chan:
		fmt.Println(value)
	case value := <-string_chan:
		panic(value)
	}
}
```
