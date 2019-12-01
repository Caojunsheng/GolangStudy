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

### 3.golang 字符串拼接及相关benchmark测试结果
golang中字符串拼接有多种方式：
1. 使用"+"来拼接字符串
2. 使用bytes.buffer来拼接
3. 使用fmt.Sprintf来拼接
4. 使用"+="来拼接
5. 使用"strings.Join"来拼接
6. 使用"strings.Builder"来拼接

refer to [contactstring.go](./contactstring.go)

Benchmark测试结果如下：

看测试结果似乎第一种使用"+"拼接字符串最快最好，"fmt.Sprintf()"最慢，但这只是简单的两个字符串拼接，具体还是要根据实际应用场景去测试一下。
```
go test -bench="."

BenchmarkUsePlusOperator-8               1602025               743 ns/op
BenchmarkUseBytesBuffer-8                1489225               804 ns/op
BenchmarkUsePlusEqualOperator-8          1560426               767 ns/op
BenchmarkUseJoinFunction-8               1509745               799 ns/op
BenchmarkUseSprintf-8                    1265932               918 ns/op
BenchmarkUseStringBuilder-8              1506157               791 ns/op

go test -bench="." -run =^$ -cpu 1,2,4,8
goos: darwin
goarch: amd64
BenchmarkUsePlusOperator                 1592030               752 ns/op
BenchmarkUsePlusOperator-2               1618995               747 ns/op
BenchmarkUsePlusOperator-4               1561308               784 ns/op
BenchmarkUsePlusOperator-8               1545866               747 ns/op
BenchmarkUseBytesBuffer                  1459882               824 ns/op
BenchmarkUseBytesBuffer-2                1484331               801 ns/op
BenchmarkUseBytesBuffer-4                1489326               804 ns/op
BenchmarkUseBytesBuffer-8                1488888               804 ns/op
BenchmarkUsePlusEqualOperator            1536147               778 ns/op
BenchmarkUsePlusEqualOperator-2          1564294               767 ns/op
BenchmarkUsePlusEqualOperator-4          1562044               768 ns/op
BenchmarkUsePlusEqualOperator-8          1561665               770 ns/op
BenchmarkUseJoinFunction                 1578355               759 ns/op
BenchmarkUseJoinFunction-2               1564411               774 ns/op
BenchmarkUseJoinFunction-4               1573408               763 ns/op
BenchmarkUseJoinFunction-8               1570466               767 ns/op
BenchmarkUseSprintf                      1334482               939 ns/op
BenchmarkUseSprintf-2                    1307694               930 ns/op
BenchmarkUseSprintf-4                    1278780               909 ns/op
BenchmarkUseSprintf-8                    1315984               899 ns/op
BenchmarkUseStringBuilder                1497134               810 ns/op
BenchmarkUseStringBuilder-2              1518432               791 ns/op
BenchmarkUseStringBuilder-4              1515824               791 ns/op
BenchmarkUseStringBuilder-8              1511204               792 ns/op

```