package main

import (
	"fmt"
	"iter"
)

func fibonacciSequence(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		a, b := 0, 1

		for range n {
			if !yield(a) {
				return
			}

			a, b = b, a+b
		}
	}
}

func main() {
	for fib := range fibonacciSequence(10) {
		fmt.Printf("num: %d\n", fib)
	}
}
