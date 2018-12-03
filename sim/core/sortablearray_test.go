package core

import (
	"math"
	"testing"
)

// Test object for testing
type SortableObject struct {
	distance float32
}

func (s *SortableObject) GetDistance() float32 {
	return s.distance
}

func floatCompare(a, b float32) bool {
	return math.Abs(float64(a-b)) < 0.001
}

func TestEmptySA(t *testing.T) {
	sa := NewSortableArray(3)
	if len(sa.Items) != 0 {
		t.Error("No items should exist")
	}
}

func TestSingleSA(t *testing.T) {
	sa := NewSortableArray(3)
	sa.Add(&SortableObject{distance: 1.0})

	if len(sa.Items) != 1 || !floatCompare(1.0, sa.Items[0].GetDistance()) {
		t.Error("Single item addition failure")
	}
}

func TestNoConflictSA(t *testing.T) {
	sa := NewSortableArray(3)
	sa.Add(&SortableObject{distance: 1.0})
	sa.Add(&SortableObject{distance: 2.0})
	sa.Add(&SortableObject{distance: 3.0})
	sa.Add(&SortableObject{distance: 4.0})

	if len(sa.Items) != 3 {
		t.Error("Should'nt have more than 3 elements in SA")
	}

	if !floatCompare(1.0, sa.Items[0].GetDistance()) {
		t.Error("1 should be in 1 spot.")
	}

	if !floatCompare(2.0, sa.Items[1].GetDistance()) {
		t.Error("2 should be in 2 spot.")
	}

	if !floatCompare(3.0, sa.Items[2].GetDistance()) {
		t.Error("3 should be in 3 spot.")
	}
}

func TestReverseOrderedSA(t *testing.T) {
	sa := NewSortableArray(3)
	sa.Add(&SortableObject{distance: 4.0})
	sa.Add(&SortableObject{distance: 3.0})
	sa.Add(&SortableObject{distance: 2.0})
	sa.Add(&SortableObject{distance: 1.0})

	if len(sa.Items) != 3 {
		t.Error("Should'nt have more than 3 elements in SA")
	}

	if !floatCompare(1.0, sa.Items[0].GetDistance()) {
		t.Error("1 should be in 1 spot.")
	}

	if !floatCompare(2.0, sa.Items[1].GetDistance()) {
		t.Error("2 should be in 2 spot.")
	}

	if !floatCompare(3.0, sa.Items[2].GetDistance()) {
		t.Error("3 should be in 3 spot.")
	}
}
