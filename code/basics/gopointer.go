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
