package editorEngine

import (
	"fmt"
	"go-simulate-a-city/sim/input"

	"github.com/go-gl/glfw/v3.2/glfw"
)

type EditorEngine struct {
	engineModeRegs     []chan EditorMode
	engineAddModeRegs  []chan EditorAddMode
	engineDrawModeRegs []chan EditorDrawMode

	engineState              State
	keyPressChannel          chan glfw.Key
	EngineModeRegChannel     chan chan EditorMode
	EngineAddModeRegChannel  chan chan EditorAddMode
	EngineDrawModeRegChannel chan chan EditorDrawMode
	ControlChannel           chan int
}

func NewEditorEngine(keyPressRegChannel chan chan glfw.Key) *EditorEngine {
	engine := EditorEngine{
		engineState: State{
			Mode:             Select,
			InAddMode:        PowerPlant,
			InDrawMode:       TerrainFlatten,
			ItemSubSelection: Item1,
			SnapToGrid:       true,
			SnapToElements:   true,
			SnapToAngle:      false},
		keyPressChannel:          make(chan glfw.Key, 2),
		engineModeRegs:           make([]chan EditorMode, 0),
		engineAddModeRegs:        make([]chan EditorAddMode, 0),
		engineDrawModeRegs:       make([]chan EditorDrawMode, 0),
		EngineModeRegChannel:     make(chan chan EditorMode),
		EngineAddModeRegChannel:  make(chan chan EditorAddMode),
		EngineDrawModeRegChannel: make(chan chan EditorDrawMode),
		ControlChannel:           make(chan int)}

	keyPressRegChannel <- engine.keyPressChannel

	go engine.run()
	return &engine
}

func (e *EditorEngine) run() {
	for {
		select {
		case reg := <-e.EngineModeRegChannel:
			e.engineModeRegs = append(e.engineModeRegs, reg)
			break
		case reg := <-e.EngineAddModeRegChannel:
			e.engineAddModeRegs = append(e.engineAddModeRegs, reg)
			break
		case reg := <-e.EngineDrawModeRegChannel:
			e.engineDrawModeRegs = append(e.engineDrawModeRegs, reg)
			break
		case key := <-e.keyPressChannel:
			// updated => used to avoid duplicate checks.
			updated := e.checkEditorMode(key)
			updated = updated || e.checkEditorToggles(key)
			if e.engineState.Mode == Add {
				updated = updated || e.checkAddMode(key)
				updated = updated || e.checkAddModeSubSelections(key)
			} else if e.engineState.Mode == Draw {
				updated = updated || e.checkDrawModeSubSelections(key)
			}
			break
		case _ = <-e.ControlChannel:
			return
		}
	}
}

func (e *EditorEngine) checkEditorMode(key glfw.Key) bool {
	selectionChanged := false
	switch key {
	case input.GetKeyCode(input.SelectModeKey):
		e.engineState.Mode = Select
		fmt.Println("Entered selection mode.")
		selectionChanged = true
	case input.GetKeyCode(input.AddModeKey):
		e.engineState.Mode = Add
		fmt.Println("Entered addition mode.")
		selectionChanged = true

		for _, reg := range e.engineAddModeRegs {
			reg <- e.engineState.InAddMode
		}
	case input.GetKeyCode(input.DrawModeKey):
		e.engineState.Mode = Draw
		fmt.Println("Entered draw mode.")
		selectionChanged = true

		for _, reg := range e.engineDrawModeRegs {
			reg <- e.engineState.InDrawMode
		}
	default:
	}

	if selectionChanged {
		for _, reg := range e.engineModeRegs {
			reg <- e.engineState.Mode
		}
	}

	return selectionChanged
}

func (e *EditorEngine) checkEditorToggles(key glfw.Key) bool {
	switch key {
	case input.GetKeyCode(input.SnapToGridKey):
		e.engineState.SnapToGrid = !e.engineState.SnapToGrid
		fmt.Printf("Toggled snap-to-grid to %v.\n", e.engineState.SnapToGrid)
		return true
	case input.GetKeyCode(input.SnapToAngleKey):
		e.engineState.SnapToAngle = !e.engineState.SnapToAngle
		fmt.Printf("Toggled snap-to-angle to %v.\n", e.engineState.SnapToAngle)
		return true
	case input.GetKeyCode(input.SnapToElementsKey):
		e.engineState.SnapToElements = !e.engineState.SnapToElements
		fmt.Printf("Toggled snap-to-elements to %v.\n", e.engineState.SnapToElements)
		return true
	default:
		return false
	}
}

func (e *EditorEngine) checkAddMode(key glfw.Key) bool {
	selectionChanged := false
	switch key {
	case input.GetKeyCode(input.PowerPlantAddModeKey):
		e.engineState.InAddMode = PowerPlant
		fmt.Println("Entered powerplant add mode.")
		selectionChanged = true
	case input.GetKeyCode(input.PowerLineAddModeKey):
		e.engineState.InAddMode = PowerLine
		fmt.Println("Entered powerline add mode.")
		selectionChanged = true
	case input.GetKeyCode(input.RoadLineAddModeKey):
		e.engineState.InAddMode = RoadLine
		fmt.Println("Entered roadline add mode.")
		selectionChanged = true
	default:
	}

	if selectionChanged {
		for _, reg := range e.engineAddModeRegs {
			reg <- e.engineState.InAddMode
		}
	}

	return selectionChanged
}

func (e *EditorEngine) checkAddModeSubSelections(key glfw.Key) bool {
	switch key {
	case input.GetKeyCode(input.ItemAdd1Key):
		e.engineState.ItemSubSelection = Item1
		fmt.Println("Selected sub-selection 1")
		return true
	case input.GetKeyCode(input.ItemAdd2Key):
		e.engineState.ItemSubSelection = Item2
		fmt.Println("Selected sub-selection 2")
		return true
	case input.GetKeyCode(input.ItemAdd3Key):
		e.engineState.ItemSubSelection = Item3
		fmt.Println("Selected sub-selection 3")
		return true
	case input.GetKeyCode(input.ItemAdd4Key):
		e.engineState.ItemSubSelection = Item4
		fmt.Println("Selected sub-selection 4")
		return true
	case input.GetKeyCode(input.ItemAdd5Key):
		e.engineState.ItemSubSelection = Item5
		fmt.Println("Selected sub-selection 5")
		return true
	case input.GetKeyCode(input.ItemAdd6Key):
		e.engineState.ItemSubSelection = Item6
		fmt.Println("Selected sub-selection 6")
		return true
	default:
		return false
	}
}

func (e *EditorEngine) checkDrawModeSubSelections(key glfw.Key) bool {
	selectionChanged := false
	switch key {
	case input.GetKeyCode(input.TerrainFlattenKey):
		e.engineState.InDrawMode = TerrainFlatten
		fmt.Println("Selected terrain flatten tool")
		selectionChanged = true
	case input.GetKeyCode(input.TerrainSharpenKey):
		e.engineState.InDrawMode = TerrainSharpen
		fmt.Println("Selected terrain sharpen tool")
		selectionChanged = true
	case input.GetKeyCode(input.TerrainTreesKey):
		e.engineState.InDrawMode = TerrainTrees
		fmt.Println("Selected terrain trees tool")
		selectionChanged = true
	case input.GetKeyCode(input.TerrainShrubsKey):
		e.engineState.InDrawMode = TerrainShrubs
		fmt.Println("Selected terrain shrubs tool")
		selectionChanged = true
	case input.GetKeyCode(input.TerrainHillsKey):
		e.engineState.InDrawMode = TerrainHills
		fmt.Println("Selected terrain hills tool")
		selectionChanged = true
	case input.GetKeyCode(input.TerrainValleysKey):
		e.engineState.InDrawMode = TerrainValleys
		fmt.Println("Selected terrain valleys tool")
		selectionChanged = true
	default:
	}

	if selectionChanged {
		for _, reg := range e.engineDrawModeRegs {
			reg <- e.engineState.InDrawMode
		}
	}

	return selectionChanged
}
