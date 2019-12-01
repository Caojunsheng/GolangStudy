package main

import (
	"math/rand"
	"testing"
)
var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func BenchmarkUsePlusOperator(t *testing.B) {
	for i := 0; i < t.N; i++ {
		usePlusOperator(randSeq(10), randSeq(10))
	}
}

func BenchmarkUseBytesBuffer(t *testing.B) {
	for i := 0; i < t.N; i++ {
		useBytesBuffer(randSeq(10), randSeq(10))
	}
}

func BenchmarkUsePlusEqualOperator(t *testing.B) {
	for i := 0; i < t.N; i++ {
		usePlusEqualOperator(randSeq(10), randSeq(10))
	}
}

func BenchmarkUseJoinFunction(t *testing.B) {
	for i := 0; i < t.N; i++ {
		useJoinFunction(randSeq(10), randSeq(10))
	}
}

func BenchmarkUseSprintf(t *testing.B) {
	for i := 0; i < t.N; i++ {
		useSprintf(randSeq(10), randSeq(10))
	}
}

func BenchmarkUseStringBuilder(t *testing.B) {
	for i := 0; i < t.N; i++ {
		useStringsBuilder(randSeq(10), randSeq(10))
	}
}
