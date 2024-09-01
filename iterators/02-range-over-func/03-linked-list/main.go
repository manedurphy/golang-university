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
