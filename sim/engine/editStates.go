package engine

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Defines the edit state for segments where there may be a first element
type EditState struct {
	hasFirstNode     bool
	firstNode        mgl32.Vec2
	firstNodeElement int64
}

func NewEditState() *EditState {
	s := EditState{}
	s.Reset()
	return &s
}

func (p *EditState) Reset() {
	p.hasFirstNode = false
}
