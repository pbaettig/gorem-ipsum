package fifo

import (
	"sync"
)

// Int ...
type Int struct {
	sync.RWMutex
	index    int
	data     []int
	capacity int
}

// Add ...
func (i *Int) Add(n int) {
	i.Lock()
	defer i.Unlock()

	if len(i.data) >= i.capacity {
		// max capacity reached
		newdata := make([]int, i.capacity-1)
		copy(newdata, i.data[1:])
		i.data = newdata
	}

	i.data = append(i.data, n)
	i.index++

}

// Get ..
func (i *Int) Get() []int {
	i.RLock()
	defer i.RUnlock()

	return i.data
}

// NewInt ..
func NewInt(cap int) *Int {
	i := new(Int)
	i.capacity = cap
	i.data = make([]int, 0)

	return i
}
