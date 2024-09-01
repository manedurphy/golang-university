---
layout: post
title: Generators With Golang Iterators
tags: golang iterators generators
---

# Table of Contents

- [Table of Contents](#table-of-contents)
- [What is a Generator?](#what-is-a-generator)
- [Example 1: Number Generator](#example-1-number-generator)
	- [Channels](#channels)
		- [Basic](#basic)
		- [Leaking Goroutine](#leaking-goroutine)
		- [Control Channel](#control-channel)
	- [Iterators](#iterators)
- [Example 2: Prime Number Generator](#example-2-prime-number-generator)
- [Example 3: Fibonacci Sequence](#example-3-fibonacci-sequence)
- [Example 4: Memory Efficiency](#example-4-memory-efficiency)
	- [Slices](#slices)
	- [Iterators](#iterators-1)
- [Conclusion](#conclusion)

# What is a Generator?

In programming, a generator is a type of iterator that allows you to iterate over a set of data without storing the entire dataset in memory. This approach is much more memory-efficient than a function that returns an array of all computed values. We will demonstrate this concept in Golang using the new [iterator](https://tip.golang.org/wiki/RangefuncExperiment) feature introduced experimentally in version [1.22](https://tip.golang.org/doc/go1.22) and officially released in version [1.23](https://tip.golang.org/doc/go1.23).

# Example 1: Number Generator

## Channels

### Basic

Let's look at a basic number generator that uses channels. In this example, we expect numbers from `20` to `25` to be sent through the channel returned by the `numbersGenChan` function, printed to the console, and then the channel to close.


```go
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
	}
}
```

We can see from the output that this implementation fulfills the requirements of a generator because the consumer operates on one value at a time, and the entire dataset is not stored in memory.

```txt
yielding number to consumer: 20
number was received by consumer

yielding number to consumer: 21
number received in range-loop: 20
number received in range-loop: 21
number was received by consumer

yielding number to consumer: 22
number was received by consumer

yielding number to consumer: 23
number received in range-loop: 22
number received in range-loop: 23
number was received by consumer

yielding number to consumer: 24
number was received by consumer

yielding number to consumer: 25
number received in range-loop: 24
number received in range-loop: 25
number was received by consumer

closing channel
```

### Leaking Goroutine

What happens if we want to break out of our loop early? This would cause a leaking goroutine because the channel never closes!

```go
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
```

We can see from the output that the log for the closed channel is missing.

```txt
yielding number to consumer: 20
number was received by consumer

yielding number to consumer: 21
number received in range-loop: 20
number received in range-loop: 21
number was received by consumer

yielding number to consumer: 22
number was received by consumer

yielding number to consumer: 23
number received in range-loop: 22
number received in range-loop: 23
```

### Control Channel

To address the leaking goroutine from the previous example, we need to ensure that the goroutine in `generateNumbers` always returns, even on a `break`. We can achieve this with a control channel.

```go
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
```

We can see from the output that the log for the closed channel is restored.

```txt
yielding number to consumer: 20
number was received by consumer

yielding number to consumer: 21
number received in range-loop: 20
number received in range-loop: 21
number was received by consumer

yielding number to consumer: 22
number was received by consumer

yielding number to consumer: 23
number received in range-loop: 22
number received in range-loop: 23
number was received by consumer

yielding number to consumer: 24
closing channel
```

If you are an experienced Golang developer, the code we've explored should be straightforward. However, other programming and scripting languages, such as Python, provide this functionality out of the box via the `yield` keyword. How can we achieve the same thing with Golang without having to build a solution ourselves with goroutines? The answer is iterators!

## Iterators

Let's examine how the same number generator is built with the new iterator feature in Golang `1.23`. Essentially, the value passed into the `yield` function is the value that will be seen in our `for-range` loop. We can see that `yield` returns a boolean, which is `false` when the loop reaches a `break` statement.

```go
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
```

We can see from the output that `yield` returns `false` once the value `23` is encountered. This is due to the `break` statement in the `for-range` loop.

```txt
yielding number to consumer: 20
number received in range-loop: 20
number was received by consumer

yielding number to consumer: 21
number received in range-loop: 21
number was received by consumer

yielding number to consumer: 22
number received in range-loop: 22
number was received by consumer

yielding number to consumer: 23
number received in range-loop: 23
stopping now
```

All examples moving forward will only use the new iterator feature.

# Example 2: Prime Number Generator

Let's look at another example where the generator returns only prime numbers. From this example, we should expect the program to end once the first prime number greater than `20` is encountered.

The interesting part of this example is the infinite loop that is started in the `generatePrimeNumbers` function. Does this mean that the scope of that infinite `for` loop will run for the duration of the entire program? No! The loop's scope will only run once per iteration in `main`. The moment we reach the break statement is when `yield` will return false and stop the infinite loop.

```go
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
```

We can see from the output that the program ends once a prime number greater than `20` is encountered.

```txt
num: 2
num: 3
num: 5
num: 7
num: 11
num: 13
num: 17
num: 19
num: 23
```

# Example 3: Fibonacci Sequence

Let's see how we can use iterators to get the nth digit of the Fibonacci sequence. Similar to how our previous example maintained the state of `n`, we are maintaining the state of variables `a` and `b`.

```go
package main

import (
	"fmt"
	"iter"
)

func fibonacciSequence(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		a, b := 0, 1

		for range n {
			if !yield(a) {
				return
			}

			a, b = b, a+b
		}
	}
}

func main() {
	for fib := range fibonacciSequence(10) {
		fmt.Printf("num: %d\n", fib)
	}
}
```

We can see from the output that the last digit is the 10th digit in the Fibonacci sequence.

```txt
num: 0
num: 1
num: 1
num: 2
num: 3
num: 5
num: 8
num: 13
num: 21
num: 34
```

# Example 4: Memory Efficiency

Let's examine the memory efficiency of generators.

## Slices

In this example, we will create `Course` objects, store them in a slice, perform an operation on each element in the slice, and look at the memory allocated.

```go
package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

type Course struct {
	ID         int
	Name       string
	University string
}

var (
	courseNames = []string{
		"Chem-1",
		"Chem-2",
		"Physics-1",
		"Physics-2",
		"Physics-3",
		"Calculus-1",
		"Calculus-2",
		"Calculus-3",
	}

	universities = []string{
		"SJSU",
		"SDSU",
		"UCB",
		"UCSF",
	}
)

func generateCourses(numCourses int) []Course {
	var courses []Course

	for i := range numCourses {
		courses = append(courses, Course{
			ID:         i,
			Name:       courseNames[rand.Intn(len(courseNames))],
			University: universities[rand.Intn(len(universities))],
		})
	}

	return courses
}

func main() {
	// Memory before generating courses
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	now := time.Now()
	courses := generateCourses(10000000)
	fmt.Printf("took %.2f seconds to generate slice\n", time.Since(now).Seconds())

	now = time.Now()
	for _, course := range courses {
		course.ID++
	}
	fmt.Printf("took %.2f seconds to operate on all courses\n", time.Since(now).Seconds())

	// Memory after generating courses
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	fmt.Printf("total allocated memory (before): %.2f Mb\n", float64(memBefore.TotalAlloc)/1e6)
	fmt.Printf("total allocated memory (after): %.2f Mb\n", float64(memAfter.TotalAlloc)/1e6)
}
```

We can see from the output that 2 gigabytes of memory were allocated when generating a slice of courses. How does this compare with using iterators?

```txt
took 1.45 seconds to generate slice
took 0.06 seconds to operate on all courses
total allocated memory (before): 0.34 Mb
total allocated memory (after): 2020.16 Mb
```

## Iterators

In this example, we will create `Course` objects via a generator, perform an operation on each element, and look at the memory allocated.

```go
package main

import (
	"fmt"
	"iter"
	"math/rand"
	"runtime"
	"time"
)

type Course struct {
	ID         int
	Name       string
	University string
}

var (
	courseNames = []string{
		"Chem-1",
		"Chem-2",
		"Physics-1",
		"Physics-2",
		"Physics-3",
		"Calculus-1",
		"Calculus-2",
		"Calculus-3",
	}

	universities = []string{
		"SJSU",
		"SDSU",
		"UCB",
		"UCSF",
	}
)

func generateCourses(numCourses int) iter.Seq[Course] {
	return func(yield func(Course) bool) {
		for i := range numCourses {
			course := Course{
				ID:         i,
				Name:       courseNames[rand.Intn(len(courseNames))],
				University: universities[rand.Intn(len(universities))],
			}

			if !yield(course) {
				return
			}
		}
	}
}

func main() {
	// Memory before generating courses
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	now := time.Now()
	courses := generateCourses(10000000)
	since := time.Since(now)
	fmt.Printf("took %.2f seconds to create iterator\n", since.Seconds())

	now = time.Now()
	for course := range courses {
		course.ID++
	}
	since = time.Since(now)
	fmt.Printf("took %.2f seconds to operate on all courses\n", since.Seconds())

	// Memory after generating courses
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	fmt.Printf("total allocated memory (before): %.2f Mb\n", float64(memBefore.TotalAlloc)/1e6)
	fmt.Printf("total allocated memory (after): %.2f Mb\n", float64(memAfter.TotalAlloc)/1e6)
}
```

We can see from the output that the generator has no effect on memory allocation. This is because we are no longer storing all those objects in a container in memory before operating on them. Instead, we create a `Course` object and operate on it before creating and operating on the next one. We see a slight increase in the time it takes to operate on each object, due to the fact that the object has to be created on each iteration.

```txt
took 0.00 seconds to generate sequence
took 0.23 seconds to operate on all courses
total allocated memory (before): 0.34 Mb
total allocated memory (after): 0.34 Mb
```  

# Conclusion

By leveraging the new iterator feature introduced in Golang `1.23`, we simplified generator implementation, making code cleaner and more efficient. Our examples showed that iterators not only provide a concise way to handle sequences but also offer significant advantages in memory efficiency. While slices consume considerable memory as they store all elements at once, iterators generate values on-the-fly, which helps in managing large datasets without unnecessary memory overhead.
