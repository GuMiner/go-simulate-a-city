package engine

import (
	"fmt"
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/engine/core"
	"go-simulate-a-city/sim/engine/core/dto"
	"go-simulate-a-city/sim/engine/element"
	"go-simulate-a-city/sim/engine/power"
	"go-simulate-a-city/sim/engine/road"
	"go-simulate-a-city/sim/engine/terrain"
	"go-simulate-a-city/sim/input/editorEngine"

	"github.com/go-gl/mathgl/mgl32"
)

type Engine struct {
	terrainMap          *terrain.TerrainMap
	elementFinder       *element.ElementFinder
	powerGrid           *power.PowerGrid
	roadGrid            *road.RoadGrid
	infiniRoadGenerator *road.InfiniRoadGenerator

	isMousePressed  bool
	actionPerformed bool
	lastBoardPos    mgl32.Vec2
	powerLineState  *PowerLineEditState
	roadLineState   *RoadLineEditState
	snapElements    SnapElements

	Hypotheticals HypotheticalActions
}

// TODO: This should not take in these parameters. These should be populated in the
// well-known listings of channels
func NewEngine(
	camOffsetRegChannel chan chan mgl32.Vec2,
	camScaleRegChannel chan chan float32) *Engine {
	terrain.Init(config.Config.Terrain.Generation.Seed)

	engine := Engine{}

	engine.terrainMap = terrain.NewTerrainMap(camOffsetRegChannel, camScaleRegChannel)
	engine.elementFinder = element.NewElementFinder()
	engine.powerGrid = power.NewPowerGrid()
	engine.roadGrid = road.NewRoadGrid()
	engine.infiniRoadGenerator = road.NewInfiniRoadGenerator(
		engine.roadGrid,
		engine.elementFinder,
		engine.terrainMap.NewRegionRegChannel)
	engine.isMousePressed = false
	engine.actionPerformed = false
	engine.powerLineState = NewPowerLineEditState()
	engine.roadLineState = NewRoadLineEditState()
	engine.snapElements = NewSnapElements()

	engine.Hypotheticals = NewHypotheticalActions()
	return &engine
}

func (e *Engine) addPowerPlantIfValid() {
	intesectsWithElement := e.elementFinder.IntersectsWithElement(e.getEffectivePosition(), e.Hypotheticals.Regions[0].Region.Scale)

	if !intesectsWithElement {
		isGroundValid := e.terrainMap.ValidateGroundLocation(e.Hypotheticals.Regions[0].Region)
		if isGroundValid {
			plantType := power.GetPlantType(editorEngine.Item1) // TODO: EngineState.ItemSubSelection)
			plantSize := power.Small                            // TODO: Configurable

			element := e.powerGrid.Add(e.getEffectivePosition(), plantType, plantSize)
			e.elementFinder.Add(element)
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
		line := e.powerGrid.AddLine(e.powerLineState.firstNode,
			powerLineEnd, 1000,
			e.powerLineState.firstNodeElement, e.getEffectivePowerGridElement())
		if line != nil {
			e.elementFinder.Add(line)
			powerLineCost := e.powerLineState.firstNode.Sub(powerLineEnd).Len() * config.Config.Power.PowerLineCost
			core.CoreFinances.TransactionChannel <- dto.NewTransaction("Power Line", powerLineCost)

			e.powerLineState.firstNode = powerLineEnd
			e.powerLineState.firstNodeElement = line.GetSnapNodeElement(1)
		}
	}
}

func (e *Engine) updateRoadLineState() {
	// TODO: Deduplicate
	// If this is the first press, we associate it with the first location of the powerline.
	if !e.roadLineState.hasFirstNode {
		e.roadLineState.firstNode = e.getEffectivePosition()
		e.roadLineState.hasFirstNode = true
		e.roadLineState.firstNodeElement = e.getEffectiveRoadGridElement()
	} else {
		// TODO: Configurable capacity
		roadLineEnd := e.getEffectivePosition()
		line := e.roadGrid.AddLine(e.roadLineState.firstNode,
			roadLineEnd, 1000,
			e.roadLineState.firstNodeElement, e.getEffectiveRoadGridElement())
		if line != nil {
			e.elementFinder.Add(line)
			roadLineCost := e.roadLineState.firstNode.Sub(roadLineEnd).Len() * 3000 // TODO: Configurable
			core.CoreFinances.TransactionChannel <- dto.NewTransaction("Road", roadLineCost)

			e.roadLineState.firstNode = roadLineEnd
			e.roadLineState.firstNodeElement = line.GetSnapNodeElement(1)
		}
	}
}

func (e *Engine) getEffectivePosition() mgl32.Vec2 {
	if e.snapElements.snappedNode != nil {
		return e.snapElements.snappedNode.Element.GetSnapNodes()[e.snapElements.snappedNode.SnapNodeIdx]
	}

	if e.snapElements.snappedGridPos != nil {
		return *e.snapElements.snappedGridPos
	}

	return e.lastBoardPos
}

// TODO: Rename, element is too generic...
func (e *Engine) getEffectivePowerGridElement() int {
	node := e.snapElements.snappedNode
	if node != nil {
		// TODO: New interface for power elements?
		if line, ok := node.Element.(*power.PowerLine); ok {
			return line.GetSnapNodeElement(node.SnapNodeIdx)
		}

		if powerPlant, ok := node.Element.(*power.PowerPlant); ok {
			return powerPlant.GetSnapElement()
		}

		panic(fmt.Sprintf("We've snapped to a node that isn't a power grid element: %v\n", node))
	}

	// No grid element association.
	return -1
}

func (e *Engine) getEffectiveRoadGridElement() int {
	node := e.snapElements.snappedNode
	if node != nil {
		if line, ok := node.Element.(*road.RoadLine); ok {
			return line.GetSnapNodeElement(node.SnapNodeIdx)
		}

		panic(fmt.Sprintf("We've snapped to a node that isn't a road element: %v\n", node))
	}

	// No grid element association.
	return -1
}

func (e *Engine) MousePress(pos mgl32.Vec2, engineState editorEngine.State) {
	e.isMousePressed = true
	e.lastBoardPos = pos
	if !e.actionPerformed {
		if engineState.Mode == editorEngine.Add && engineState.InAddMode == editorEngine.PowerPlant {
			e.addPowerPlantIfValid()
		}
	}
}

func (e *Engine) MouseMoved(pos mgl32.Vec2, engineState editorEngine.State) {
	e.lastBoardPos = pos

	e.powerLineState.EnterOrExitEditMode(&engineState)
	e.roadLineState.EnterOrExitEditMode(&engineState)
}

func (e *Engine) MouseRelease(pos mgl32.Vec2, engineState editorEngine.State) {
	e.isMousePressed = false
	e.actionPerformed = false
	e.lastBoardPos = pos

	if e.powerLineState.InPowerLineState(&engineState) {
		e.updatePowerLineState()
	}

	if e.roadLineState.InRoadLineState(&engineState) {
		e.updateRoadLineState()
	}
}

// Cancels the state of any multi-step operation, resetting it back to the start.
func (e *Engine) CancelState(engineState editorEngine.State) {
	if e.powerLineState.InPowerLineState(&engineState) {
		e.powerLineState.Reset()
	}

	if e.roadLineState.InRoadLineState(&engineState) {
		e.roadLineState.Reset()
	}
}

func (e *Engine) applyStepDraw(stepAmount float32, engineState *editorEngine.State) {
	region := e.Hypotheticals.Regions[0].Region
	stepFactor := 0.1 * stepAmount

	switch engineState.InDrawMode {
	case editorEngine.TerrainFlatten:
		e.terrainMap.Flatten(region, stepFactor)
	case editorEngine.TerrainSharpen:
		e.terrainMap.Sharpen(region, stepFactor)
	case editorEngine.TerrainHills:
		e.terrainMap.Hills(region, stepFactor)
	case editorEngine.TerrainValleys:
		e.terrainMap.Valleys(region, stepFactor)
	default:
		break
	}
}

// Performs operations that are performed as steps with time for edit
func (e *Engine) StepEdit(stepAmount float32, engineState editorEngine.State) {
	if engineState.Mode == editorEngine.Draw && e.isMousePressed {
		e.applyStepDraw(stepAmount, &engineState)
	}
}

// Update methods based on UI
func (e *Engine) PrecacheRegions(regions []commonMath.IntVec2) {
	for _, region := range regions {
		e.terrainMap.GetOrAddRegion(region.X(), region.Y())
	}
}

func (e *Engine) ComputeSnapNodes(engineState *editorEngine.State) {
	e.snapElements.ComputeSnappedSnapElements(e.lastBoardPos, e.elementFinder, engineState)
}

func (e *Engine) GetTerrainMap() *terrain.TerrainMap {
	return e.terrainMap
}

func (e *Engine) GetPowerGrid() *power.PowerGrid {
	return e.powerGrid
}

func (e *Engine) GetRoadGrid() *road.RoadGrid {
	return e.roadGrid
}

func (e *Engine) GetElementFinder() *element.ElementFinder {
	return e.elementFinder
}

func (e *Engine) GetSnapElements() *SnapElements {
	return &e.snapElements
}
