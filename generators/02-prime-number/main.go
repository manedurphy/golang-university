package main

import (
	"fmt"
	"iter"
	"math"
)

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}

	sqrtN := int(math.Sqrt(float64(n)))
	for i := 2; i <= sqrtN; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func generatePrimeNumbers() iter.Seq[int] {
	return func(yield func(i int) bool) {
		n := 0

		for {
			if isPrime(n) {
				if !yield(n) {
					return
				}
			}

			n++
		}
	}
}

func main() {
	for num := range generatePrimeNumbers() {
		fmt.Printf("prime number received: %d\n", num)

		if num > 20 {
			break
		}
	}
}
