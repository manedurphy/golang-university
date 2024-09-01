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
