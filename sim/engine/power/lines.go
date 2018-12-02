package power

import (
	"go-simulate-a-city/common/commonmath"
)

type PowerTerminus struct {
}

type PowerLine struct {
	capacity int64
}

// Implement Element
// Gets the central position of the element.
func (p *PowerLine) GetRegion() *commonMath.Region {
	return nil
}

// Gets positions on the map that can be used to snap to points of the element.
// func (p *PowerLine) GetSnapNodes() []mgl32.Vec2 {
// 	return []mgl32.Vec2{
// 		p.start,
// 		p.end}
// }

// Gets lines on the map that can be used to snap to *edges* of the element
// func (p *PowerLine) GetSnapEdges() [][2]mgl32.Vec2 {
// 	return [][2]mgl32.Vec2{
// 		[2]mgl32.Vec2{
// 			p.start,
// 			p.end}}
// }
//
// // Gets the line this power line represents.
// func (p *PowerLine) GetLine() [2]mgl32.Vec2 {
// 	return p.GetSnapEdges()[0]
// }

// func (p *PowerLine) GetSnapNodeElement(snapNodeIdx int) int {
// 	if snapNodeIdx == 0 {
// 		return p.startNode
// 	}
//
// 	return p.endNode
// }
