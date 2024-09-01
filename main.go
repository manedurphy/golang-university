package main

import (
	"fmt"
)

func getNumbers(done chan struct{}) <-chan int {
	ch := make(chan int)

	go func() {
		defer func() {
			fmt.Println("closing channel...")
			close(ch)
		}()

		for _, val := range []int{1, 2, 3, 4, 5, 6, 7, 8} {
			select {
			case <-done:
				fmt.Println("can you seen me!")
				done <- struct{}{}
				return
			case ch <- val:
			}
		}
	}()

	return ch
}

func main() {
	done := make(chan struct{})
	defer func() {
		fmt.Println("cleaning up...")
		done <- struct{}{}
		<-done
	}()

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		fmt.Println("recovered from panic:", r)
	// 	}
	// }()

	for val := range getNumbers(done) {
		fmt.Printf("value: %d\n", val)

		if val == 3 {
			panic("panicking in for-range loop!")
		}
	}
}

// func mayPanic() {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			fmt.Println("Recovered from panic in mayPanic:", r)
// 		}
// 	}()

// 	defer func() {
// 		fmt.Println("secondary defer")
// 	}()

// 	// Simulate a panic
// 	panic("an unexpected error occurred")
// }
