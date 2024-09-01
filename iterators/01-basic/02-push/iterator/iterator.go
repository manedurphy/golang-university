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
