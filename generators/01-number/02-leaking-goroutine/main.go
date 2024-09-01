package main

import "fmt"

func generateNumbers() <-chan int {
	ch := make(chan int)

	go func() {
		for i := 20; i <= 25; i++ {
			fmt.Printf("yielding number to consumer: %d\n", i)
			ch <- i

			fmt.Println("number was received by consumer")
			fmt.Println()
		}

		fmt.Println("closing channel")
		close(ch)
	}()

	return ch
}

func main() {
	for num := range generateNumbers() {
		fmt.Printf("number received in range-loop: %d\n", num)

		if num == 23 {
			break
		}
	}
}
