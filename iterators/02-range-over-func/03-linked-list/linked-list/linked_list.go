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
