package cmap

import "sync"

// Defines an int64-indexed simple cooncurrent map.
// Storing nil entries is undefined.
type Map struct {
	data          map[int64]interface{}
	lock          sync.Mutex
	nextItemIndex int64
}

func NewMap() *Map {
	return &Map{
		data:          make(map[int64]interface{}),
		nextItemIndex: 0}
}

// Iteratively adds an item to the map, returning the index of the item
func (m *Map) IterativeAdd(item interface{}) int64 {
	if item == nil {
		panic("Storing nil items is disallowed!")
	}

	m.lock.Lock()
	defer m.lock.Unlock()

	itemIndex := m.nextItemIndex
	m.nextItemIndex++
	m.data[itemIndex] = item

	return itemIndex
}

// Gets an item
func (m *Map) Get(index int64) interface{} {
	m.lock.Lock()
	defer m.lock.Unlock()

	if data, ok := m.data[index]; ok {
		return data
	}

	return nil
}

// Stores an item
func (m *Map) Set(index int64, item interface{}) {
	if item == nil {
		panic("Storing nil items is disallowed!")
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.data[index] = item
}
