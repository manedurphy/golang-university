package main

import (
	"fmt"

	"github.com/manedurphy/golang-university/iterators/01-basic/01-pull/iterator"
)

func main() {
	it := iterator.NewIterator()

	for {
		val, ok := it.Next()
		if !ok {
			fmt.Println("no more values")
			break
		}

		fmt.Printf("value: %d\n", val)
	}
}
