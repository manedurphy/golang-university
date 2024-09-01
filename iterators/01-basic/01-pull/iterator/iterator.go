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
