package main

import (
	"context"
	"fmt"

	"github.com/manedurphy/golang-university/iterators/01-basic/02-push/iterator"
)

func main() {
	it := iterator.NewIterator()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for val := range it.GetNumbers(ctx) {
		fmt.Printf("value: %d\n", val)

		if val == 45 {
			cancel()
			break
		}
	}

	fmt.Println("no more values")
}
