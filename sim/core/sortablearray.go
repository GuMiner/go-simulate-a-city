package core

type Sortable interface {
	GetDistance() float32
}

// Defines a self-sorting array (ascending order)
type SortableArray struct {
	maxSize int
	Items   []Sortable
}

func NewSortableArray(maxSize int) *SortableArray {
	return &SortableArray{maxSize: maxSize, Items: make([]Sortable, 0)}
}

func (s *SortableArray) moveItemDown(idx int) {
	// Don't move it down if we'd be extending our array past the max size
	if idx >= s.maxSize-1 {
		return
	}

	if idx == len(s.Items)-1 {
		// Add a new item if we're moving it down off the edge of the array
		s.Items = append(s.Items, s.Items[idx])
	} else {
		// Modify the assignment otherwise
		s.Items[idx+1] = s.Items[idx]
	}
}

// Adds an item into the array, which is then sorted appropriately or dropped
func (s *SortableArray) Add(item Sortable) {
	itemLen := len(s.Items)
	addedItem := false
	for currentIndex := len(s.Items); currentIndex > 0; currentIndex-- {
		if s.Items[currentIndex-1].GetDistance() > item.GetDistance() {
			s.moveItemDown(currentIndex - 1)
			s.Items[currentIndex-1] = item
			addedItem = true
		} else {
			break
		}
	}

	// Array isn't full and this is the last item, so add it.
	if !addedItem && itemLen < s.maxSize {
		s.Items = append(s.Items, item)
	}
}
