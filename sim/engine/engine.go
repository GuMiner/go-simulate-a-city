package engine

import (
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/core/dto/editorengdto"
	"go-simulate-a-city/sim/core/mailroom"
	"go-simulate-a-city/sim/engine/core"
	"go-simulate-a-city/sim/engine/core/dto"
	"go-simulate-a-city/sim/engine/finder"
	"go-simulate-a-city/sim/engine/power"
	"go-simulate-a-city/sim/engine/road"
	"go-simulate-a-city/sim/engine/terrain"
	"go-simulate-a-city/sim/input/editorEngine"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

type Engine struct {
	terrainMap          *terrain.TerrainMap
	elementFinder       *finder.ElementFinder
	powerGrid           *power.PowerGrid
	roadGrid            *road.RoadGrid
	infiniRoadGenerator *road.InfiniRoadGenerator

	isMousePressed bool
	lastBoardPos   mgl32.Vec2
	powerLineState *EditState
	roadLineState  *EditState
	snap           *Snap

	editorMode     editorengdto.EditorMode
	editorAddMode  editorengdto.EditorAddMode
	editorDrawMode editorengdto.EditorDrawMode

	editorModeChannel     chan editorengdto.EditorMode
	editorAddModeChannel  chan editorengdto.EditorAddMode
	editorDrawModeChannel chan editorengdto.EditorDrawMode

	Hypotheticals HypotheticalActions

	mousePressChannel    chan glfw.MouseButton
	mouseReleaseChannel  chan glfw.MouseButton
	mouseBoardPosChannel chan mgl32.Vec2
	ControlChannel       chan int
}

func NewEngine() *Engine {
	terrain.Init(config.Config.Terrain.Generation.Seed)

	engine := Engine{
		editorMode:            editorengdto.Select,
		editorAddMode:         editorengdto.PowerPlant,
		editorDrawMode:        editorengdto.TerrainFlatten,
		editorModeChannel:     make(chan editorengdto.EditorMode),
		editorAddModeChannel:  make(chan editorengdto.EditorAddMode),
		editorDrawModeChannel: make(chan editorengdto.EditorDrawMode),
		mouseBoardPosChannel:  make(chan mgl32.Vec2, 10),
		mousePressChannel:     make(chan glfw.MouseButton, 10),
		mouseReleaseChannel:   make(chan glfw.MouseButton, 10),
		ControlChannel:        make(chan int)}

	engine.terrainMap = terrain.NewTerrainMap()
	mailroom.NewTerrainRegChannel = engine.terrainMap.NewTerrainRegChannel
	mailroom.NewRegionRegChannel = engine.terrainMap.NewRegionRegChannel

	engine.elementFinder = finder.NewElementFinder()
	engine.powerGrid = power.NewPowerGrid()
	engine.roadGrid = road.NewRoadGrid()
	engine.infiniRoadGenerator = road.NewInfiniRoadGenerator(engine.roadGrid)
	engine.isMousePressed = false
	engine.powerLineState = NewEditState()
	engine.roadLineState = NewEditState()
	engine.snap = NewSnap(engine.elementFinder)

	engine.Hypotheticals = NewHypotheticalActions()

	mailroom.MousePressedRegChannel <- engine.mousePressChannel
	mailroom.MouseReleasedRegChannel <- engine.mouseReleaseChannel
	mailroom.BoardPosChangeRegChannel <- engine.mouseBoardPosChannel
	mailroom.EngineModeRegChannel <- engine.editorModeChannel
	mailroom.EngineAddModeRegChannel <- engine.editorAddModeChannel
	mailroom.EngineDrawModeRegChannel <- engine.editorDrawModeChannel

	go engine.run()
	return &engine
}

// func (e *Engine) updateHypotheticalsAndSnapNodes() {
// 	e.Hypotheticals.ComputeHypotheticalRegion(engine, &editorEngine.EngineState)
// }

func (e *Engine) run() {
	for {
		select {
		case e.lastBoardPos = <-e.mouseBoardPosChannel:
		case e.editorMode = <-e.editorModeChannel:
		case e.editorAddMode = <-e.editorAddModeChannel:
		case e.editorDrawMode = <-e.editorDrawModeChannel:

		case _ = <-e.mousePressChannel:
			e.isMousePressed = true

			if e.editorMode == editorengdto.Add && e.editorAddMode == editorengdto.PowerPlant {
				e.addPowerPlantIfValid()
			}
		case _ = <-e.mouseReleaseChannel:
			e.isMousePressed = false

			if e.editorMode == editorengdto.Add {
				if e.editorAddMode == editorengdto.PowerLine {
					e.updatePowerLineState()
				} else if e.editorAddMode == editorengdto.RoadLine {
					e.updateRoadLineState()
				}
			}
		case _ = <-e.ControlChannel:
			return
		}
	}
}

func (e *Engine) addPowerPlantIfValid() {
	intesectsWithElement := false // e.elementFinder.IntersectsWithElement(e.getEffectivePosition(), e.Hypotheticals.Regions[0].Region.Scale)

	if !intesectsWithElement {
		isGroundValid := true // e.terrainMap.ValidateGroundLocation(e.Hypotheticals.Regions[0].Region)
		if isGroundValid {
			plantType := power.GetPlantType(editorengdto.Item1) // TODO: EngineState.ItemSubSelection)
			plantSize := power.Small                            // TODO: Configurable

			_ = e.powerGrid.Add(e.getEffectivePosition(), plantType, plantSize)
			// e.elementFinder.Add(element)
			core.CoreFinances.TransactionChannel <- dto.NewTransaction("Power Plant", power.GetPlantCost(plantType))
		}
	}
}

func (e *Engine) updatePowerLineState() {
	// If this is the first press, we associate it with the first location of the powerline.
	if !e.powerLineState.hasFirstNode {
		e.powerLineState.firstNode = e.getEffectivePosition()
		e.powerLineState.hasFirstNode = true
		e.powerLineState.firstNodeElement = e.getEffectivePowerGridElement()
	} else {
		// TODO: Configurable capacity
		powerLineEnd := e.getEffectivePosition()
		lineId := e.powerGrid.AddLine(e.powerLineState.firstNode,
			powerLineEnd, 1000,
			e.powerLineState.firstNodeElement, e.getEffectivePowerGridElement())
		if lineId != -1 {
			// e.elementFinder.Add(line)
			powerLineCost := e.powerLineState.firstNode.Sub(powerLineEnd).Len() * config.Config.Power.PowerLineCost
			core.CoreFinances.TransactionChannel <- dto.NewTransaction("Power Line", powerLineCost)

			e.powerLineState.firstNode = powerLineEnd
			e.powerLineState.firstNodeElement = lineId
		}
	}
}

func (e *Engine) updateRoadLineState() {
	// TODO: Deduplicate
	// If this is the first press, we associate it with the first location of the powerline.
	if !e.roadLineState.hasFirstNode {
		e.roadLineState.firstNode = e.getEffectivePosition()
		e.roadLineState.hasFirstNode = true
		e.roadLineState.firstNodeElement = int64(e.getEffectiveRoadGridElement())
	} else {
		// TODO: Configurable capacity
		roadLineEnd := e.getEffectivePosition()
		lineId := e.roadGrid.AddLine(e.roadLineState.firstNode,
			roadLineEnd, 1000,
			e.roadLineState.firstNodeElement, int64(e.getEffectiveRoadGridElement()))
		if lineId != -1 {
			// e.elementFinder.Add(line)
			roadLineCost := e.roadLineState.firstNode.Sub(roadLineEnd).Len() * 3000 // TODO: Configurable
			core.CoreFinances.TransactionChannel <- dto.NewTransaction("Road", roadLineCost)

			e.roadLineState.firstNode = roadLineEnd
			e.roadLineState.firstNodeElement = lineId
		}
	}
}

func (e *Engine) getEffectivePosition() mgl32.Vec2 {
	// if e.snapElements.snappedNode != nil {
	// 	return e.snapElements.snappedNode.Element.GetSnapNodes()[e.snapElements.snappedNode.SnapNodeIdx]
	// }
	//
	// if e.snapElements.snappedGridPos != nil {
	// 	return *e.snapElements.snappedGridPos
	// }

	return e.lastBoardPos
}

// TODO: Rename, element is too generic...
func (e *Engine) getEffectivePowerGridElement() int64 {
	// node := e.snapElements.snappedNode
	// if node != nil {
	// 	// TODO: New interface for power elements?
	// 	// if line, ok := node.Element.(*power.PowerLine); ok {
	// 	// 	return line.GetSnapNodeElement(node.SnapNodeIdx)
	// 	// }
	// 	//
	// 	// if powerPlant, ok := node.Element.(*power.PowerPlant); ok {
	// 	// 	return powerPlant.GetSnapElement()
	// 	// }
	// 	//
	// 	// panic(fmt.Sprintf("We've snapped to a node that isn't a power grid element: %v\n", node))
	// }

	// No grid element association.
	return -1
}

func (e *Engine) getEffectiveRoadGridElement() int {
	// node := nil // e.snapElements.snappedNode
	// if node != nil {
	// 	if line, ok := node.Element.(*road.RoadLine); ok {
	// 		return line.GetSnapNodeElement(node.SnapNodeIdx)
	// 	}
	//
	// 	panic(fmt.Sprintf("We've snapped to a node that isn't a road element: %v\n", node))
	// }

	// No grid element association.
	return -1
}

// Cancels the state of any multi-step operation, resetting it back to the start.
func (e *Engine) CancelState(engineState editorEngine.State) {
	// if e.powerLineState.InPowerLineState(&engineState) {
	// 	e.powerLineState.Reset()
	// }
	//
	// if e.roadLineState.InRoadLineState(&engineState) {
	// 	e.roadLineState.Reset()
	// }
}

func (e *Engine) applyStepDraw(stepAmount float32, engineState *editorEngine.State) {
	region := e.Hypotheticals.Regions[0].Region
	stepFactor := 0.1 * stepAmount

	switch engineState.InDrawMode {
	case editorengdto.TerrainFlatten:
		e.terrainMap.Flatten(region, stepFactor)
	case editorengdto.TerrainSharpen:
		e.terrainMap.Sharpen(region, stepFactor)
	case editorengdto.TerrainHills:
		e.terrainMap.Hills(region, stepFactor)
	case editorengdto.TerrainValleys:
		e.terrainMap.Valleys(region, stepFactor)
	default:
		break
	}
}

// Performs operations that are performed as steps with time for edit
func (e *Engine) StepEdit(stepAmount float32, engineState editorEngine.State) {
	if engineState.Mode == editorengdto.Draw && e.isMousePressed {
		e.applyStepDraw(stepAmount, &engineState)
	}
}
