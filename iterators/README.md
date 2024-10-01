# Table of Contents
- [Table of Contents](#table-of-contents)
- [What Are Iterators?](#what-are-iterators)
- [Example 1: Basic](#example-1-basic)
	- [Pull](#pull)
		- [Iterator Package](#iterator-package)
		- [Consumer](#consumer)
	- [Push](#push)
		- [Iterator Package](#iterator-package-1)
		- [Consumer](#consumer-1)
- [Example 2: Range Over Func](#example-2-range-over-func)
	- [Basic](#basic)
	- [Iterator Revised](#iterator-revised)
	- [Linked List](#linked-list)
- [Example 3: Deep Dive](#example-3-deep-dive)
	- [Sequence Of Events](#sequence-of-events)
	- [Defer Statements](#defer-statements)
	- [Panic](#panic)
		- [Iterator](#iterator)
		- [Loop Body](#loop-body)
		- [Pull](#pull-1)
- [Example 4: Database](#example-4-database)
	- [Database](#database)
		- [Push](#push-1)
		- [Pull](#pull-2)

# What Are Iterators?

In programming, an iterator can be defined as an abstract entity which provides access to a collection's data one element at a time. The consumer of an iterator does not need to be aware of the structure of the data which is maintained by the iterator, it just has to understand that it has sequential access to each element in the iterator's underlying data structure. We will look at how to build iterators from scratch in Go, and then explore the new `range-over-function` iterators that were officially released as part of the [1.23](https://go.dev/doc/go1.23) release of Golang.

# Example 1: Basic

Before looking at the new `range-over-function` iterators, let's build a `pull` and `push` iterator from scratch to become more familiar with the concept. 

## Pull

The `pull` iterator approach provides a straightforward way to iterate over the elements of the underlying data structure, and provides the control of starting and stopping the iteration to the consumer. When the consumer calls the `Next` method, it is requesting the iterator to provide the next element. While control of the iteration may be advantageous in some cases, that responsibility may be a disadvantage sometimes as the consumer's code may become more complex with having to both process the data it is requesting, as well as be mindful of the state that the iterator is maintaining internally.

### Iterator Package

For this example, the iterator is an interface with one method called `Next`. `Next` returns an integer value and a boolean, which is `false` when the iteration over the underlying data structure ends. The underlying data structure, in this case, is a slice.

The constructor creates a hardcoded slice of integers as the underlying data structure. This is to make the example easy to understand. Remember, an iterator is simply an abstraction which allows the consumer to have sequential access to its values. The consumer does not know what the underlying data structure of the iterator is. The `Next` method returns the value in the slice that is at the index being tracked by the `idx` field. This field is incremented each time `Next` is called. When the value of `idx` is greater than or equal to the length of the slice, we know that the iteration has completed and return `false` to the caller.


```go
package iterator

type (
	Iterator interface {
		// Next returns the next sequential value and a boolean which
		// indicates if the value it valid. When there are no more values,
		// a zero is returned for the value and the boolean is "false".
		Next() (val int, ok bool)
	}

	iterator struct {
		idx  int
		data []int
	}
)

func NewIterator() Iterator {
	return &iterator{
		idx:  0,
		data: []int{3, 2, 45, 4, 6, 7},
	}
}

func (i *iterator) Next() (int, bool) {
	if i.idx >= len(i.data) {
		return 0, false
	}

	val := i.data[i.idx]
	i.idx++

	return val, true
}
```

### Consumer

In our main function, we are instantiating a new instance of an iterator and using an infinite `for-loop` to continuously call the `Next` method. For each call to `Next`, the boolean return value is evaluated to determine if the iteration is complete. Each value is printed to the console.

```go
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
```

## Push 

One of the many advantages of `push` iterators is that it is suitable for event-driven systems where data is generated asynchronously. Imagine that the consumer in a Go program is a `goroutine` that simply listens on a channel; waiting for new data to be produced such that it can respond accordingly. A `push` iterator is more suitable for this case than a `pull` iterator. The consumer code is also simpler than `pull` iterators because it is not responsible for controlling the iteration, nor does it need to evaluate the status of the iteration.

### Iterator Package

With a `push` model, it is the iterator's responsibility to feed the elements of its underlying data structure to the consumer sequentially. We can replace the `Next` method with a `GetNumbers` method which will send each element of the underlying slice through a channel. Since we do not need to track the index with this implementation, the `idx` field can be removed.

With this implementation, we can see that the consumer will be able to iterate over the elements with a `for-range` loop. However, this code is at risk of leaking a `goroutine`. If the consumer decides to `break` out of the `for-range` loop early, the channel will be blocked. Let's force the consumer to provide a context to this method.

```go
package iterator

type (
	Iterator interface {
		// GetNumbers returns a channel for sequential access to all numbers
		// in the underlying data structure
		GetNumbers() <-chan int
	}

	iterator struct {
		data []int
	}
)

func NewIterator() Iterator {
	return &iterator{
		data: []int{3, 2, 45, 4, 6, 7},
	}
}

func (i *iterator) GetNumbers() <-chan int {
	ch := make(chan int)

	go func() {
		defer close(ch)

		for _, val := range i.data {
			ch <- val
		}
	}()

	return ch
}
```

To ensure that our `goroutine` does not leak, the consumer will need to cancel the context that it passes to the `GetNumbers` method if it wants to `break` out of its `for-range` loop early.

```go
package iterator

import "context"

type (
	Iterator interface {
		// GetNumbers returns a channel for sequential access to all numbers
		// in the underlying data structure
		GetNumbers(ctx context.Context) <-chan int
	}

	iterator struct {
		data []int
	}
)

func NewIterator() Iterator {
	return &iterator{
		data: []int{3, 2, 45, 4, 6, 7},
	}
}

func (i *iterator) GetNumbers(ctx context.Context) <-chan int {
	ch := make(chan int)

	go func() {
		defer close(ch)

		for _, val := range i.data {
			select {
			case <-ctx.Done():
				return
			default:
				ch <- val
			}
		}
	}()

	return ch
}
```

### Consumer

In our main function, we are iterating over the channel that is returned by the `GetNumbers` method in a `for-range` loop. We have provided the method with the context that it needs to close the channel if the consumer breaks out of its iteration early. In this case, the consumer breaks out of the `for-range` loop when the value 45 is detected. The consumer is responsible for cancelling the context.

The issue with this implementation is that the consumer of a `push` iterator should not bare any responsibility for the internal workings of the iterator. The consumer has essentially taken on the responsibility of closing the iterator's channel. This is not ideal when dealing with an abstraction. The new `range-over-function` iterators solve this problem.

```go
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
```

# Example 2: Range Over Func

## Basic

From the official [documentation](https://go.dev/wiki/RangefuncExperiment), we can see that a new `iter` package was added to the standard library as part of an experiemental feature in Go version `1.22`. This feature is now available in `1.23`. There are two new generic types defined in this package, `Seq` and `Seq2`. The `Seq` type allows us to get sequential access to a single value, and the `Seq2` type allows us to get access to two values. Let's look at a basic example.

The `getNumbers` function returns an iterator of type `int` using the `Seq` type. Since `Seq` is generic, we must specify the type that we intend to work with. Within the scope of the `getNumbers` function, we are iterating through a hardcoded slice and passing each element to the `yield` function. Like the previous example, a hardcoded slice is used to make the example simple. We will see a more practical example later.

```go
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
```

If you are familiar with other programming and scripting languages like Python, you might recall that `yield` is a keyword which facilitates the creation of generator functions, and allows for lazy evaluation where you generate values as needed rather than all upfront. In Go, `yield` is no a keyword, but rather a naming convention for a function which essentiall achieves the same functionality. When `yield` is called on the integer value, it is passed to the consumer of the iterator immediately for processing. We must check the return value of `yield` to ensure that we stop iterating when it returns `false`. The `yield` function returns `false` when the loop-body of the consumer calls `break`. We will see that in the next section. When you run this program, you will see that each value in the slice is printed to the console.

## Iterator Revised

In this revised example of the `iterator` package, the `GetNumbers` method returns an iterator instead of a channel. The consumer is not longer responsible for closing a channel, and still has sequential access to each element of the underlying data structure without knowing what that data structure looks like.

```go
package iterator

import "iter"

type (
	Iterator interface {
		// GetNumbers returns an iterator for sequential access to all numbers
		// in the underlying data structure
		GetNumbers() iter.Seq[int]
	}

	iterator struct {
		data []int
	}
)

func NewIterator() Iterator {
	return &iterator{
		data: []int{3, 2, 45, 4, 6, 7},
	}
}

func (i *iterator) GetNumbers() iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, val := range i.data {
			if !yield(val) {
				return
			}
		}
	}
}
```

In our main function, we can now iterate using a `for-range` loop without having to provide a context to the iterator.

```go
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
```

## Linked List

As mentioned already, an iterator's underlying data structure can be anything. To demonstrate this, we can use a linked list which provides a `Traverse` method to traverse each node. Since the `Seq` type is generic, we can specify that we are working with pointers to `Node` types.

In the `Traverse` method, we start from the head of the list, and iterate using the internal `next` field to update the value of the `current` variable. When `current` is `nil`, the iteration ends. The consumer can also end the iteration early via the `break` keyword since the iterator is properly checking the return value for the `yield` function.

```go
package linked_list

import "iter"

type (
	LinkedList struct {
		head *Node
	}

	Node struct {
		value int
		next  *Node
	}
)

func NewLinkedList() *LinkedList {
	return &LinkedList{}
}

// Append adds a new node with the specified value to the end of the linked list
func (ll *LinkedList) Append(value int) {
	newNode := Node{value: value}
	if ll.head == nil {
		ll.head = &newNode
		return
	}

	current := ll.head
	for current.next != nil {
		current = current.next
	}
	current.next = &newNode
}

// Traverse returns an iterator for sequential access to all nodes in the linked list
func (ll *LinkedList) Traverse() iter.Seq[*Node] {
	return func(yield func(*Node) bool) {
		current := ll.head
		for current != nil {
			if !yield(current) {
				return
			}

			current = current.next
		}
	}
}
```

In our main function, we populate the linked list with several values before iterating through each node via the `Traverse` method in a `for-range` loop.

```go
package main

import (
	"fmt"

	linked_list "github.com/manedurphy/golang-university/iterators/02-range-over-func/03-linked-list/linked-list"
)

func main() {
	linkedList := linked_list.NewLinkedList()

	linkedList.Append(3)
	linkedList.Append(2)
	linkedList.Append(45)
	linkedList.Append(4)
	linkedList.Append(6)
	linkedList.Append(7)

	for node := range linkedList.Traverse() {
		fmt.Printf("node: %+v\n", node)
	}
}
```

# Example 3: Deep Dive

## Sequence Of Events

This is an example that is similar to what we've seen before with a number of logs added to to show us what happens. In the `getNumbers` function, we see a value `n` which is set to `20`. We then see a conditional loop which continues so long as `n` is less than or equal to `21`. A log prints the value of `n` at the beginning of the loop's scope, a log for the case where the `yield` function returns `false`, and a log which shows the value of `n` just after it is incremented. In our main function, we are iterating over our function iterator with a `for-range` loop as we've seen in previous examples. Considering everything we have discussed so far, see if you can accurately predict the output of this program.

```go
package main

import (
	"fmt"
	"iter"
)

func getNumbers() iter.Seq[int] {
	return func(yield func(int) bool) {
		n := 20
		for n <= 21 {
			fmt.Printf("hello from iterator: n=%d\n", n)
			if !yield(n) {
				fmt.Println("stopping iteration")
				return
			}

			n++
			fmt.Printf("incrementing n: n=%d\n", n)
		}
	}
}

func main() {
	for val := range getNumbers() {
		fmt.Printf("value: %d\n", val)

		if val == 21 {
			break
		}
	}
}
```

Our iterator shows that the value of `n` is 20 as the loop begins. The next log that we can see is within the `for-range` loop in the main function. So, when a value is passed into `yield` it is essentially blocked until it receives a return value of `true` or `false`. If the `for-range` loop continues its iteration, then the return value of `yield` is `true`. We can see that's the case here because we see a subsequent log for the value of `n` after it is incremented. The value that is returned by the `yield` function is `false` when a break statement is encountered in a `for-range` loop. This is confirmed in the next iteration when `n` is `21`. The value is passed to the `for-range` loop via the `yield` function, the `for-range` loop sees that the value is `21` and breaks, and the log for the stopping of the iteration is seen back in the iterator. This example illustrates the back-and-forth execution between the function iterator and the `for-range` loop.

```
hello from iterator: n=20
value: 20
incrementing n: n=21
hello from iterator: n=21
value: 21
stopping iteration
```

## Defer Statements

This example expands the previous. The difference here is that two `defer` statements have been added; one in the iterator and one in the `for-range` loop. Based on your current understanding of `defer` statements in Go, what do you expect to happen?

The official documentation states that the semantics of `defer` do not depend on what kind of value is being ranged over. This means that we can expect the `defer` statements in the iterator to run when the iterator function returns, and we can expect the `defer` statements in the `for-range` loop to run when the main function returns.

```go
package main

import (
	"fmt"
	"iter"
)

func getNumbers() iter.Seq[int] {
	return func(yield func(int) bool) {
		n := 20
		for n <= 21 {
			defer func() {
				fmt.Println("deferred from iterator")
			}()

			fmt.Printf("hello from iterator: n=%d\n", n)
			if !yield(n) {
				fmt.Println("stopping iteration")
				return
			}

			n++
			fmt.Printf("incrementing n: n=%d\n", n)
		}
	}
}

func main() {
	for val := range getNumbers() {
		defer func() {
			fmt.Println("deferred from for-range loop body")
		}()

		fmt.Printf("value: %d\n", val)

		if val == 21 {
			break
		}
	}

	fmt.Println("exiting...")
}
```

As we can see, when our iterator returns, the deferred statement runs twice. We know that this happens when the iterator returns because we can see the log just before the return statement is printed to the console before the two logs from the deferred statement. We can see two logs at the end of the program after the log that says `exiting...`. This shows that the semantics of the defer statment in Go do not change because of the type the is ranged over.

```
hello from iterator: n=20
value: 20
incrementing n: n=21
hello from iterator: n=21
value: 21
stopping iteration
deferred from iterator
deferred from iterator
exiting...
deferred from for-range loop body
deferred from for-range loop body
```

## Panic

The promise of how panics are handled is the same as `defer` statements. There are no surprises, as the semantics have not changed for iterators.

### Iterator

This example shows what happens when a `panic` occurs in an iterator. In Go, when a `panic` occurs, the deferred statements are executed in LIFO (Last In, First Out) order when the surrounding function exits. Here, the last `defer` is inside the `for-loop` of the iterator. This deferred statement occurs twice, followed by the one at the beginning of the iterator's scope.

```go
package main

import (
	"fmt"
	"iter"
)

func getNumbers() iter.Seq[int] {
	return func(yield func(int) bool) {
		defer func() {
			fmt.Println("deferred from iterator beginning")
		}()

		n := 20
		for n <= 21 {
			defer func() {
				fmt.Println("deferred from iterator for-loop")
			}()

			fmt.Printf("hello from iterator: n=%d\n", n)
			if !yield(n) {
				fmt.Println("stopping iteration")
				return
			}

			if n == 21 {
				`panic`("panicking in iterator")
			}

			n++
			fmt.Printf("incrementing n: n=%d\n", n)
		}
	}
}

func main() {
	defer func() {
		fmt.Println("deferred from main")
	}()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered from `panic`:", r)
		}
	}()

	for val := range getNumbers() {
		defer func() {
			fmt.Println("deferred from for-range loop body")
		}()

		fmt.Printf("value: %d\n", val)
	}
}
```

We can see that the `defer` statements occur in LIFO order as we would expect them to in any Go program.

```
hello from iterator: n=20
value: 20
incrementing n: n=21
hello from iterator: n=21
value: 21
deferred from iterator for-loop
deferred from iterator for-loop
deferred from iterator beginning
deferred from for-range loop body
deferred from for-range loop body
recovered from `panic`: panicking in iterator
deferred from main
```

### Loop Body

We can see that the `defer` statements from the iterator occur first, as they were the last in the queue of `defer` statements. The `defer` statements within the main function occur after.

```
hello from iterator: n=20
value: 20
incrementing n: n=21
hello from iterator: n=21
value: 21
deferred from iterator for-loop
deferred from iterator for-loop
deferred from iterator beginning
deferred from for-range loop body
deferred from for-range loop body
recovered from `panic`: panicking in for-range loop!
deferred from main
```

### Pull

So far, each example of a Golang iterator has been a `push` iterator. This is because the iterator has controlled the tempo of each iteration, while the `for-range` loop body has simply waited for new data to be available. The `iter` package has a `Pull` function which returns two functions, `next` and `stop`. The `next` function returns the next value in the iterator's sequence as well as a boolean to indicate whether the value is valid. The boolean is `false` when the last value in the sequence has been pulled.

The call to `stop` is deferred to the end of the function, but if we call it earlier, then all calls to `next` will be invalid.

```go
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
		`panic`("not good")
	}
	fmt.Printf("num: %d\n", val)

	val, ok = next()
	if !ok {
		`panic`("not good")
	}
	fmt.Printf("num: %d\n", val)

	val, ok = next()
	if !ok {
		`panic`("not good")
	}
	fmt.Printf("num: %d\n", val)
}
```

# Example 4: Database

Let's explore a practical example where we use iterators to retrieve data from a database.

## Database

The database functionality has been encapsulated within a package named `db`, which is a common practice to enhance code readability and write unit tests. The main point of interest here is the `GetCourses` method, which utilizes the `iter.Seq2` definition to return two values to the consumer: a `Course` object and an `error`.

```go
// Seq2 is an iterator over sequences of pairs of values, most commonly key-value pairs. When called as seq(yield), seq calls yield(k, v) for each pair (k, v) in the sequence, stopping early if yield returns false.
type Seq2[K, V any] func(yield func(K, V) bool)

// GetCourses returns an iterator of Course objects
GetCourses() iter.Seq2[Course, error]
```

This method begins by querying all rows. If an error occurs, it is returned to the consumer, and the iteration stops since there is nothing to iterate through. As we process each row, note that we do not return immediately upon encountering an error; instead, we yield the error back to the caller and proceed to the next row. You can decide to continue or stop based on your application's needs. This example shows that encountering an error while scanning a row does not necessarily require stopping the iteration.

```go
package db

import (
	"database/sql"
	"fmt"
	"iter"
	"math/rand"

	_ "github.com/mattn/go-sqlite3"
)

type (
	CoursesDB interface {
		// Seed seeds the database with the number of courses specified
		Seed(numCourses int) error

		// GetCourses returns an iterator of Course objects
		GetCourses() iter.Seq2[Course, error]

		// Close closes the database
		Close() error
	}

	Course struct {
		ID         int
		Name       string
		University string
	}

	coursesDB struct {
		db *sql.DB
	}
)

const (
	selectSQL    = `SELECT * FROM courses`
	insertSQL    = `INSERT INTO courses(name, university) VALUES (?, ?)`
	dropTableSQL = `DROP TABLE IF EXISTS courses`

	createTableSQL = `CREATE TABLE IF NOT EXISTS courses (
        "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,   
        "name" TEXT,
        "university" TEXT
    );`
)

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

// New creates a new CoursesDB instance
func New(dataDir string) (CoursesDB, error) {
	var (
		db  *sql.DB
		err error
	)

	// Open database
	db, err = sql.Open("sqlite3", fmt.Sprintf("%s/courses.db", dataDir))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &coursesDB{
		db: db,
	}, nil
}

func (d *coursesDB) Seed(numCourses int) error {
	var (
		tx        *sql.Tx
		statement *sql.Stmt
		err       error
	)

	_, err = d.db.Exec(dropTableSQL)
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}

	_, err = d.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	tx, err = d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	statement, err = tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare SQL statment: %w", err)
	}
	defer statement.Close()

	// Seed database
	for course := range d.generateCourses(numCourses) {
		_, err = statement.Exec(course.Name, course.University)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to prepare SQL statment: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *coursesDB) GetCourses() iter.Seq2[Course, error] {
	return func(yield func(Course, error) bool) {
		var (
			rows *sql.Rows
			err  error
		)

		rows, err = d.db.Query(selectSQL)
		if err != nil {
			// When an error is encountered, we should yield it back to
			// the consumer an stop the iterator
			yield(Course{}, err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var c Course

			err = rows.Scan(&c.ID, &c.Name, &c.University)
			if !yield(c, err) {
				return
			}
		}

		err = rows.Err()
		if err != nil {
			// When an error is encountered, we should yield it back to
			// the consumer an stop the iterator
			yield(Course{}, err)
			return
		}
	}
}

func (d *coursesDB) Close() error {
	return d.db.Close()
}

// Generator of Course objects
func (d *coursesDB) generateCourses(numCourses int) iter.Seq[Course] {
	return func(yield func(Course) bool) {
		for range numCourses {
			course := Course{
				Name:       courseNames[rand.Intn(len(courseNames))],
				University: universities[rand.Intn(len(universities))],
			}

			if !yield(course) {
				return
			}
		}
	}
}
```

### Push

With our database abstraction layer, the main program becomes very easy to read. It simply creates a new `CoursesDB` instance, seeds it with a specified number of courses, and then queries and iterates through all the data via the iterator returned by `GetCourses`.

```go
package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/manedurphy/golang-university/iterators/01-database/db"
)

var (
	dataDir    string
	numCourses int
)

func init() {
	flag.StringVar(&dataDir, "data-dir", "", "The directory for storing the DB file")
	flag.IntVar(&numCourses, "num-courses", 0, "The number of courses to create")
}

func main() {
	var (
		coursesDB db.CoursesDB
		now       time.Time
		logger    *slog.Logger
		err       error
	)

	flag.Parse()

	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create new database instance
	coursesDB, err = db.New(dataDir)
	if err != nil {
		logger.Error("failed to create database", "err", err)
		os.Exit(1)
	}
	defer coursesDB.Close()

	now, err = time.Now(), coursesDB.Seed(numCourses)
	if err != nil {
		logger.Error("failed to seed database", "err", err)
		os.Exit(1)
	}
	logger.Info("successfully seeded database", "duration_ms", time.Since(now).Milliseconds())

	// Get courses from database using iterator
	for course, err := range coursesDB.GetCourses() {
		if err != nil {
			logger.Error("failed to get course", "err", err)
			continue
		}

		logger.Info("received course", "course", course)
	}
}
```

### Pull

We can achieve the same outcome as above with a `pull` model as well.

```go
package main

import (
	"flag"
	"iter"
	"log/slog"
	"os"
	"time"

	"github.com/manedurphy/golang-university/iterators/04-database/db"
)

var (
	dataDir    string
	numCourses int
)

func init() {
	flag.StringVar(&dataDir, "data-dir", ".", "The directory for storing the DB file")
	flag.IntVar(&numCourses, "num-courses", 0, "The number of courses to create")
}

func main() {
	var (
		coursesDB db.CoursesDB
		now       time.Time
		logger    *slog.Logger
		err       error
	)

	flag.Parse()

	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Create new database instance
	coursesDB, err = db.New(dataDir)
	if err != nil {
		logger.Error("failed to create database", "err", err)
		os.Exit(1)
	}
	defer coursesDB.Close()

	now, err = time.Now(), coursesDB.Seed(numCourses)
	if err != nil {
		logger.Error("failed to seed database", "err", err)
		os.Exit(1)
	}
	logger.Info("successfully seeded database", "duration_ms", time.Since(now).Milliseconds())

	next, stop := iter.Pull2(coursesDB.GetCourses())
	defer stop()

	// Get courses from database using iterator
	for {
		course, err, valid := next()
		if !valid {
			logger.Info("iteration has completed")
			break
		}

		if err != nil {
			logger.Error("failed to get course", "err", err)
			continue
		}

		logger.Info("received course", "course", course)
	}
}
```