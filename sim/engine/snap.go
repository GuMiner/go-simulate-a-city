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
	snappedToNode    bool
	lastSnapPosition mgl32.Vec2
	lastSnapId       int64

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

	SnapQueryChannel chan SnapQuery
}

type SnapQuery struct {
	Result chan SnapResult
}

type SnapResult struct {
	IsItemSnapped bool
	Id            int64
	Position      mgl32.Vec2
}

func NewSnap(elementFinder *finder.ElementFinder) *Snap {
	s := Snap{
		snappedToNode:         false,
		lastSnapId:            -1,
		lastSnapPosition:      mgl32.Vec2{0, 0},
		elementFinder:         elementFinder,
		mouseBoardPosChannel:  make(chan mgl32.Vec2, 10),
		editorMode:            editorengdto.Select,
		editorAddMode:         editorengdto.PowerPlant,
		editorDrawMode:        editorengdto.TerrainFlatten,
		snapToGrid:            false,
		snapToAngle:           false,
		snapToElements:        true,
		editorModeChannel:     make(chan editorengdto.EditorMode),
		editorAddModeChannel:  make(chan editorengdto.EditorAddMode),
		editorDrawModeChannel: make(chan editorengdto.EditorDrawMode),
		snapSettingsChannel:   make(chan editorengdto.SnapSetting),
		SnapQueryChannel:      make(chan SnapQuery, 3)}

	mailroom.BoardPosChangeRegChannel <- s.mouseBoardPosChannel
	mailroom.EngineModeRegChannel <- s.editorModeChannel
	mailroom.EngineAddModeRegChannel <- s.editorAddModeChannel
	mailroom.EngineDrawModeRegChannel <- s.editorDrawModeChannel
	mailroom.SnapSettingsRegChannel <- s.snapSettingsChannel

	go s.run()
	return &s
}

func (s *Snap) computeSnaps(boardPos mgl32.Vec2) {
	displayedSnappedNodes := make([]mgl32.Vec2, 0)

	s.snappedToNode = false
	if s.snapToElements && s.editorMode == editorengdto.Add &&
		s.editorAddMode == editorengdto.PowerLine || s.editorAddMode == editorengdto.RoadLine {

		itemType := finder.RoadTerminus
		if s.editorAddMode == editorengdto.PowerLine {
			itemType = finder.PowerTerminus
		}

		results := make(chan []*finder.NodeWithDistance)
		s.elementFinder.KNearestSearchChannel <- finder.NewKNNQuery(boardPos, itemType, config.Config.Draw.SnapNodeCount, results)
		elements := <-results
		for _, elem := range elements {
			if elem.Distance < config.Config.Draw.MinSnapNodeDistance {
				if !s.snappedToNode {
					s.lastSnapId = elem.Id
					s.lastSnapPosition = elem.Pos
					s.snappedToNode = true
				}

				displayedSnappedNodes = append(displayedSnappedNodes, elem.Pos)
			}
		}
	}

	if s.snapToGrid {
		snapGridResolution := float32(config.Config.Snap.SnapGridResolution)
		offsetBoardPos := boardPos.Add(mgl32.Vec2{snapGridResolution / 2, snapGridResolution / 2})
		snappedIntPosition := commonMath.IntVec2{int(offsetBoardPos.X() / snapGridResolution), int(offsetBoardPos.Y() / snapGridResolution)}
		elementPos := mgl32.Vec2{float32(snappedIntPosition.X()), float32(snappedIntPosition.Y())}.Mul(snapGridResolution)

		displayedSnappedNodes = append(displayedSnappedNodes, elementPos)
		if !s.snappedToNode {
			s.lastSnapId = -1
			s.lastSnapPosition = elementPos
			s.snappedToNode = true
		}
	}

	// Send to be rendered.
	mailroom.SnappedNodesUpdateChannel <- displayedSnappedNodes
}

func (s *Snap) run() {
	for {
		select {
		case boardPos := <-s.mouseBoardPosChannel:
			// TODO drain to the last position update
			s.computeSnaps(boardPos)
		case s.editorMode = <-s.editorModeChannel:
		case s.editorAddMode = <-s.editorAddModeChannel:
			s.snappedToNode = false
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
		case query := <-s.SnapQueryChannel:
			query.Result <- SnapResult{
				IsItemSnapped: s.snappedToNode,
				Id:            s.lastSnapId,
				Position:      s.lastSnapPosition}
			close(query.Result)
		}
	}
}
