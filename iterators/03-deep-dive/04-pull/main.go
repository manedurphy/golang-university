package main

import (
	"fmt"
	"iter"
)

func getNumbers() iter.Seq[int] {
	return func(yield func(int) bool) {
		n := 0

		for {
			if !yield(n) {
				fmt.Println("done iterating!")
				return
			}

			n++
		}
	}
}

func main() {
	numbers := getNumbers()

	next, stop := iter.Pull(numbers)
	defer stop()

	val, ok := next()
	if !ok {
		panic("not good")
	}
	fmt.Printf("num: %d\n", val)

	val, ok = next()
	if !ok {
		panic("not good")
	}
	fmt.Printf("num: %d\n", val)

	val, ok = next()
	if !ok {
		panic("not good")
	}
	fmt.Printf("num: %d\n", val)
}
