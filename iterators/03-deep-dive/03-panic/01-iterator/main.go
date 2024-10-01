package main

import (
	"fmt"
	"iter"
)

func getNumbers() iter.Seq[int] {
	return func(yield func(int) bool) {
		defer func() {
			fmt.Println("deferred from iterator beginning")
		}()

		n := 20
		for n <= 21 {
			defer func() {
				fmt.Println("deferred from iterator for-loop")
			}()

			fmt.Printf("hello from iterator: n=%d\n", n)
			if !yield(n) {
				fmt.Println("stopping iteration")
				return
			}

			if n == 21 {
				panic("panicking in iterator")
			}

			n++
			fmt.Printf("incrementing n: n=%d\n", n)
		}
	}
}

func main() {
	defer func() {
		fmt.Println("deferred from main")
	}()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from panic:", r)
		}
	}()

	for val := range getNumbers() {
		defer func() {
			fmt.Println("deferred from for-range loop body")
		}()

		fmt.Printf("value: %d\n", val)
	}
}
