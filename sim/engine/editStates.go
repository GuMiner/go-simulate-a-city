package engine

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Defines the edit state for segments where there may be a first element
type EditState struct {
	wasInEditState bool

	hasFirstNode     bool
	firstNode        mgl32.Vec2
	firstNodeElement int
}

func NewEditState() *EditState {
	s := EditState{}
	s.Reset()
	return &s
}

func (p *EditState) PerformStateTransition(isInState bool) {
	if isInState && !p.wasInEditState {
		p.Reset()
		p.wasInEditState = true
	} else if !isInState && p.wasInEditState {
		p.wasInEditState = false
	}
}

func (p *EditState) Reset() {
	p.wasInEditState = false
	p.hasFirstNode = false
}
