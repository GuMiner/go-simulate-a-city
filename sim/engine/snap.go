package engine

import (
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/core/dto/editorengdto"
	"go-simulate-a-city/sim/core/mailroom"
	"go-simulate-a-city/sim/engine/finder"

	"github.com/go-gl/mathgl/mgl32"
)

type Snap struct {
	elementFinder *finder.ElementFinder

	mouseBoardPosChannel chan mgl32.Vec2

	editorMode     editorengdto.EditorMode
	editorAddMode  editorengdto.EditorAddMode
	editorDrawMode editorengdto.EditorDrawMode

	snapToGrid     bool
	snapToAngle    bool
	snapToElements bool

	editorModeChannel     chan editorengdto.EditorMode
	editorAddModeChannel  chan editorengdto.EditorAddMode
	editorDrawModeChannel chan editorengdto.EditorDrawMode
	snapSettingsChannel   chan editorengdto.SnapSetting
}

func NewSnap(elementFinder *finder.ElementFinder) Snap {
	s := Snap{
		mouseBoardPosChannel:  make(chan mgl32.Vec2, 10),
		editorMode:            editorengdto.Select,
		editorAddMode:         editorengdto.PowerPlant,
		editorDrawMode:        editorengdto.TerrainFlatten,
		editorModeChannel:     make(chan editorengdto.EditorMode),
		editorAddModeChannel:  make(chan editorengdto.EditorAddMode),
		editorDrawModeChannel: make(chan editorengdto.EditorDrawMode),
		snapSettingsChannel:   make(chan editorengdto.SnapSetting)}

	mailroom.BoardPosChangeRegChannel <- engine.mouseBoardPosChannel
	mailroom.EngineModeRegChannel <- engine.editorModeChannel
	mailroom.EngineAddModeRegChannel <- engine.editorAddModeChannel
	mailroom.EngineDrawModeRegChannel <- engine.editorDrawModeChannel
	mailroom.SnapSettingsRegChannel <- engine.snapSettingsChannel

	go s.run()
	return &s
}

func (s *Snap) computeSnaps(boardPos mgl32.Vec2) {
	if s.snapToElements && s.editorMode == editorengdto.Add &&
		s.editorAddMode == editorengdto.PowerLine || s.editorAddMode == editorengdto.RoadLine {

		itemType := finder.RoadLine
		if s.editorAddMode == editorengdto.PowerLine {
			itemType = finder.PowerLine
		}

		results := make(chan []*finder.NodeWithDistance)
		s.elementFinder.KNearestSearchChannel <- finder.NewKNNQuery(boardPos, itemType, config.Config.Draw.SnapNodeCount, results)
		elements := <-results
		for _, elem := range elements {
			if elem.Distance < config.Config.Draw.MinSnapNodeDistance {
				// TODO: we snapped to this node. Send this out for rendering / systemic use (TODO)
			}
		}
	}

	if s.snapToGrid {
		snapGridResolution := float32(config.Config.Snap.SnapGridResolution)
		offsetBoardPos := boardPos.Add(mgl32.Vec2{snapGridResolution / 2, snapGridResolution / 2})
		snappedIntPosition := commonMath.IntVec2{int(offsetBoardPos.X() / snapGridResolution), int(offsetBoardPos.Y() / snapGridResolution)}
		snappedPos := mgl32.Vec2{float32(snappedIntPosition.X()), float32(snappedIntPosition.Y())}.Mul(snapGridResolution)

		// TODO: Return the singular snapped grid pos.
	}
}

func (s *Snap) run() {
	for {
		select {
		case boardPos := <-s.mouseBoardPosChannel:
			// TODO drain to the last position update
			s.computeSnaps(boardPos)
		case s.editorMode = <-s.editorModeChannel:
		case s.editorAddMode = <-s.editorAddModeChannel:
		case s.editorDrawMode = <-s.editorDrawModeChannel:
		case snapSetting := <-s.snapSettingsChannel:
			switch snapSetting.Setting {
			case editorengdto.SnapToAngle:
				s.snapToAngle = snapSetting.State
			case editorengdto.SnapToGrid:
				s.snapToGrid = snapSetting.State
			case editorengdto.SnapToElements:
				s.snapToElements = snapSetting.State
			default:
			}
		}
	}
}
