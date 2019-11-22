### 1. golang UT coverage
When you want to scan all go files in a package, and want to calculate the ut coverage of this package

**Installation**

Get the necessary packages and its dependencies:

`$ go get github.com/axw/gocov/gocov`

`$ go get -u gopkg.in/matm/v1/gocov-html`


**Usage**

`$ gocov test ./... | gocov-html > result.html`

**Reference**

https://github.com/axw/gocov

https://github.com/matm/gocov-html

### 2. golang pointer address
When you need to change a value of a variable, but don't have direct access to it, only exposure a function to change it.
```
package main
import "fmt"
var str []string
func main() {
    setVal(&str)
    fmt.Println(str)
}
//需要在这里赋值str，但是又不能直接引用 str
func setVal(val *[]string)  {
    *val = []string{"a", "b"}
}
```
