package main

import (
	"bytes"
	"fmt"
	"strings"
)

// use + operator to contact two string
func usePlusOperator(str1, str2 string) string {
	return str1 + str2
}

// use byte.buffer with WriteString function
func useBytesBuffer(str1, str2 string) string {
	var str bytes.Buffer
	str.WriteString(str1)
	str.WriteString(str2)
	return str.String()
}

// use fmt Sprintf function
func useSprintf(str1, str2 string) string {
	str3 := fmt.Sprintf("%s%s", str1, str2)
	return str3
}

// use +=
func usePlusEqualOperator(str1, str2 string) string {
	str1 += str2
	return str1
}

// use strings.Join function
func useJoinFunction(str1, str2 string) string {
	var strs []string
	strs = append(strs, str1)
	result := strings.Join(strs, str2)
	return result
}

// use strings.Builder
func useStringsBuilder(str1, str2 string) string {
	var sb strings.Builder
	sb.WriteString(str1)
	sb.WriteString(str2)
	return sb.String()
}
