package main

import (
	"fmt"
	"iter"
)

func getNumbers() iter.Seq[int] {
	return func(yield func(int) bool) {
		n := 20
		for n <= 21 {
			defer func() {
				fmt.Println("deferred from iterator")
			}()

			fmt.Printf("hello from iterator: n=%d\n", n)
			if !yield(n) {
				fmt.Println("stopping iteration")
				return
			}

			n++
			fmt.Printf("incrementing n: n=%d\n", n)
		}
	}
}

func main() {
	for val := range getNumbers() {
		defer func() {
			fmt.Println("deferred from for-range loop body")
		}()

		fmt.Printf("value: %d\n", val)

		if val == 21 {
			break
		}
	}

	fmt.Println("exiting...")
}
