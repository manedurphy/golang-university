package main

import "fmt"

func generateNumbers(ctrl <-chan struct{}) <-chan int {
	ch := make(chan int)

	go func() {
		defer func() {
			fmt.Println("closing channel")
			close(ch)
		}()

		for i := 20; i <= 25; i++ {
			fmt.Printf("yielding number to consumer: %d\n", i)
			select {
			case ch <- i:
				fmt.Println("number was received by consumer")
				fmt.Println()
			case <-ctrl:
				return
			}
		}
	}()

	return ch
}

func main() {
	ctrl := make(chan struct{})
	for num := range generateNumbers(ctrl) {
		fmt.Printf("number received in range-loop: %d\n", num)

		if num == 23 {
			ctrl <- struct{}{}
			break
		}
	}
}
