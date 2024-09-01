package main

import (
	"fmt"
	"iter"
)

func generateNumbers() iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := 20; i <= 25; i++ {
			fmt.Printf("yielding number to consumer: %d\n", i)
			if !yield(i) {
				fmt.Println("stopping now")
				return
			}

			fmt.Println("number was received by consumer")
			fmt.Println()
		}
	}
}

func main() {
	for num := range generateNumbers() {
		fmt.Printf("number received in range-loop: %d\n", num)

		if num == 23 {
			break
		}
	}
}
