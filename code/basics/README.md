### 1. golang pointer address
When you need to change a value of a variable, but don't have direct access to it, only exposure a function to change it.

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
The randomness of select will have some issue when more than one case in select 
are satisfied at the same time, then select will random choose a case to 
execute.

If there is no case satisfied, then will execute the default sentence.

refer to **select.go**

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
Especially when you use two schedule job, and when the two schedule jobs
 triggered at the same time, then one job will not be executed this period.
