package engine

import (
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/core/dto/editorengdto"
	"go-simulate-a-city/sim/engine/element"
	"go-simulate-a-city/sim/engine/road"
	"go-simulate-a-city/sim/input/editorEngine"

	"github.com/go-gl/mathgl/mgl32"
)

type SnapElements struct {
	snappedNode    *element.ElementWithDistance
	snappedLine    *element.LineElementWithDistance
	snappedGridPos *mgl32.Vec2
}

func NewSnapElements() SnapElements {
	s := SnapElements{}
	s.Reset()
	return s
}

func (s *SnapElements) ComputeSnappedSnapElements(boardPos mgl32.Vec2, elementFinder *element.ElementFinder, engineState *editorEngine.State) {
	s.snappedNode = nil
	if engineState.SnapToElements &&
		engineState.Mode == editorengdto.Add &&
		engineState.InAddMode == editorengdto.PowerLine || engineState.InAddMode == editorengdto.RoadLine {

		elements := elementFinder.KNearest(boardPos, config.Config.Draw.SnapNodeCount)
		for _, elem := range elements {
			if elem.Distance < config.Config.Draw.MinSnapNodeDistance {
				// TODO: This logic isn't sufficiently generic. It is enough to avoid snapping powerlines to roads and vice versa, but not sufficient for future extensibility.
				_, isRoadLine := elem.Element.(*road.RoadLine)
				if (isRoadLine && engineState.InAddMode == editorengdto.RoadLine) ||
					(!isRoadLine && engineState.InAddMode != editorengdto.RoadLine) {
					s.snappedNode = &elem
					break
				}
			}
		}
	}

	s.snappedGridPos = nil
	if engineState.SnapToGrid {
		snapGridResolution := float32(config.Config.Snap.SnapGridResolution)
		offsetBoardPos := boardPos.Add(mgl32.Vec2{snapGridResolution / 2, snapGridResolution / 2})
		snappedIntPosition := commonMath.IntVec2{int(offsetBoardPos.X() / snapGridResolution), int(offsetBoardPos.Y() / snapGridResolution)}
		snappedPos := mgl32.Vec2{float32(snappedIntPosition.X()), float32(snappedIntPosition.Y())}.Mul(snapGridResolution)
		s.snappedGridPos = &snappedPos
	}
}

func (s *SnapElements) GetSnappedNode() *element.ElementWithDistance {
	return s.snappedNode
}

func (s *SnapElements) GetSnappedLine() *element.LineElementWithDistance {
	return s.snappedLine
}

func (s *SnapElements) Reset() {
	s.snappedLine = nil
	s.snappedLine = nil
}
