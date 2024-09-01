package main

import (
	"fmt"
	"iter"
)

/*
type Seq[V any] func(yield func(V) bool) bool
type Seq2[K, V any] func(yield func(K, V) bool) bool
*/

func getNumbers() iter.Seq[int] {
	return func(yield func(int) bool) {
		data := []int{3, 2, 45, 4, 6, 7}

		for _, val := range data {
			if !yield(val) {
				return
			}
		}
	}
}

func main() {
	for val := range getNumbers() {
		fmt.Printf("value: %d\n", val)
	}

	fmt.Println("no more values")
}
