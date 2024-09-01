package main

import (
	"fmt"

	"github.com/manedurphy/golang-university/iterators/02-range-over-func/02-iterator-revised/iterator"
)

func main() {
	it := iterator.NewIterator()

	for val := range it.GetNumbers() {
		fmt.Printf("value: %d\n", val)
	}

	fmt.Println("no more values")
}
