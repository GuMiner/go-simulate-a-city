package editorEngine

import (
	"fmt"
	"go-simulate-a-city/sim/core/dto/editorengdto"
	"go-simulate-a-city/sim/input"

	"github.com/go-gl/glfw/v3.2/glfw"
)

type State struct {
	Mode             editorengdto.EditorMode
	InAddMode        editorengdto.EditorAddMode
	InDrawMode       editorengdto.EditorDrawMode
	ItemSubSelection editorengdto.ItemSubSelection

	SnapSettings map[editorengdto.SnapToggle]bool
}

type EditorEngine struct {
	engineModeRegs     []chan editorengdto.EditorMode
	engineAddModeRegs  []chan editorengdto.EditorAddMode
	engineDrawModeRegs []chan editorengdto.EditorDrawMode
	snapSettingRegs    []chan editorengdto.SnapSetting

	engineState              State
	keyPressChannel          chan glfw.Key
	EngineModeRegChannel     chan chan editorengdto.EditorMode
	EngineAddModeRegChannel  chan chan editorengdto.EditorAddMode
	EngineDrawModeRegChannel chan chan editorengdto.EditorDrawMode
	SnapSettingsRegChannel   chan chan editorengdto.SnapSetting
	ControlChannel           chan int
}

func NewEditorEngine(keyPressRegChannel chan chan glfw.Key) *EditorEngine {
	engine := EditorEngine{
		engineState: State{
			Mode:             editorengdto.Select,
			InAddMode:        editorengdto.PowerPlant,
			InDrawMode:       editorengdto.TerrainFlatten,
			ItemSubSelection: editorengdto.Item1,
			SnapSettings:     make(map[editorengdto.SnapToggle]bool)},
		keyPressChannel:          make(chan glfw.Key, 2),
		engineModeRegs:           make([]chan editorengdto.EditorMode, 0),
		engineAddModeRegs:        make([]chan editorengdto.EditorAddMode, 0),
		engineDrawModeRegs:       make([]chan editorengdto.EditorDrawMode, 0),
		snapSettingRegs:          make([]chan editorengdto.SnapSetting, 0),
		EngineModeRegChannel:     make(chan chan editorengdto.EditorMode),
		EngineAddModeRegChannel:  make(chan chan editorengdto.EditorAddMode),
		EngineDrawModeRegChannel: make(chan chan editorengdto.EditorDrawMode),
		SnapSettingsRegChannel:   make(chan chan editorengdto.SnapSetting),
		ControlChannel:           make(chan int)}

	engine.engineState.SnapSettings[editorengdto.SnapToGrid] = true
	engine.engineState.SnapSettings[editorengdto.SnapToElements] = false
	engine.engineState.SnapSettings[editorengdto.SnapToAngle] = false

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
		case reg := <-e.SnapSettingsRegChannel:
			e.snapSettingRegs = append(e.snapSettingRegs, reg)
			break
		case key := <-e.keyPressChannel:
			// updated => used to avoid duplicate checks.
			updated := e.checkEditorMode(key)
			updated = updated || e.checkEditorToggles(key)
			if e.engineState.Mode == editorengdto.Add {
				updated = updated || e.checkAddMode(key)
				updated = updated || e.checkAddModeSubSelections(key)
			} else if e.engineState.Mode == editorengdto.Draw {
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
		e.engineState.Mode = editorengdto.Select
		fmt.Println("Entered selection mode.")
		selectionChanged = true
	case input.GetKeyCode(input.AddModeKey):
		e.engineState.Mode = editorengdto.Add
		fmt.Println("Entered addition mode.")
		selectionChanged = true

		for _, reg := range e.engineAddModeRegs {
			reg <- e.engineState.InAddMode
		}
	case input.GetKeyCode(input.DrawModeKey):
		e.engineState.Mode = editorengdto.Draw
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
		state := !e.engineState.SnapSettings[editorengdto.SnapToGrid]
		e.engineState.SnapSettings[editorengdto.SnapToGrid] = state

		fmt.Printf("Toggled snap-to-grid to %v.\n", state)

		for _, reg := range e.snapSettingRegs {
			reg <- editorengdto.SnapSetting{Setting: editorengdto.SnapToGrid, State: state}
		}
		return true
	case input.GetKeyCode(input.SnapToAngleKey):
		state := !e.engineState.SnapSettings[editorengdto.SnapToAngle]
		e.engineState.SnapSettings[editorengdto.SnapToAngle] = state

		fmt.Printf("Toggled snap-to-angle to %v.\n", state)

		for _, reg := range e.snapSettingRegs {
			reg <- editorengdto.SnapSetting{Setting: editorengdto.SnapToAngle, State: state}
		}
		return true
	case input.GetKeyCode(input.SnapToElementsKey):
		state := !e.engineState.SnapSettings[editorengdto.SnapToElements]
		e.engineState.SnapSettings[editorengdto.SnapToAngle] = state

		fmt.Printf("Toggled snap-to-elements to %v.\n", state)

		for _, reg := range e.snapSettingRegs {
			reg <- editorengdto.SnapSetting{Setting: editorengdto.SnapToElements, State: state}
		}
		return true
	default:
		return false
	}
}

func (e *EditorEngine) checkAddMode(key glfw.Key) bool {
	selectionChanged := false
	switch key {
	case input.GetKeyCode(input.PowerPlantAddModeKey):
		e.engineState.InAddMode = editorengdto.PowerPlant
		fmt.Println("Entered powerplant add mode.")
		selectionChanged = true
	case input.GetKeyCode(input.PowerLineAddModeKey):
		e.engineState.InAddMode = editorengdto.PowerLine
		fmt.Println("Entered powerline add mode.")
		selectionChanged = true
	case input.GetKeyCode(input.RoadLineAddModeKey):
		e.engineState.InAddMode = editorengdto.RoadLine
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
		e.engineState.ItemSubSelection = editorengdto.Item1
		fmt.Println("Selected sub-selection 1")
		return true
	case input.GetKeyCode(input.ItemAdd2Key):
		e.engineState.ItemSubSelection = editorengdto.Item2
		fmt.Println("Selected sub-selection 2")
		return true
	case input.GetKeyCode(input.ItemAdd3Key):
		e.engineState.ItemSubSelection = editorengdto.Item3
		fmt.Println("Selected sub-selection 3")
		return true
	case input.GetKeyCode(input.ItemAdd4Key):
		e.engineState.ItemSubSelection = editorengdto.Item4
		fmt.Println("Selected sub-selection 4")
		return true
	case input.GetKeyCode(input.ItemAdd5Key):
		e.engineState.ItemSubSelection = editorengdto.Item5
		fmt.Println("Selected sub-selection 5")
		return true
	case input.GetKeyCode(input.ItemAdd6Key):
		e.engineState.ItemSubSelection = editorengdto.Item6
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
		e.engineState.InDrawMode = editorengdto.TerrainFlatten
		fmt.Println("Selected terrain flatten tool")
		selectionChanged = true
	case input.GetKeyCode(input.TerrainSharpenKey):
		e.engineState.InDrawMode = editorengdto.TerrainSharpen
		fmt.Println("Selected terrain sharpen tool")
		selectionChanged = true
	case input.GetKeyCode(input.TerrainTreesKey):
		e.engineState.InDrawMode = editorengdto.TerrainTrees
		fmt.Println("Selected terrain trees tool")
		selectionChanged = true
	case input.GetKeyCode(input.TerrainShrubsKey):
		e.engineState.InDrawMode = editorengdto.TerrainShrubs
		fmt.Println("Selected terrain shrubs tool")
		selectionChanged = true
	case input.GetKeyCode(input.TerrainHillsKey):
		e.engineState.InDrawMode = editorengdto.TerrainHills
		fmt.Println("Selected terrain hills tool")
		selectionChanged = true
	case input.GetKeyCode(input.TerrainValleysKey):
		e.engineState.InDrawMode = editorengdto.TerrainValleys
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
