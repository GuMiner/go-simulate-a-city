package ui

import (
	"go-simulate-a-city/common/commonio"
	"go-simulate-a-city/sim/input/editorEngine"

	"github.com/go-gl/glfw/v3.2/glfw"
)

type CustomCursorType int

const (
	Selection CustomCursorType = iota
	PowerPlantAdd
	PowerLineAdd
	RoadLineAdd

	TerrainFlatten
	TerrainSharpen
	TerrainTrees
	TerrainShrubs
	TerrainHills
	TerrainValleys
)

type CustomCursors struct {
	cursors         map[CustomCursorType]*glfw.Cursor
	drawModeCursors map[editorEngine.EditorDrawMode]CustomCursorType
	addModeCursors  map[editorEngine.EditorAddMode]CustomCursorType

	globalEditEngineChan chan editorEngine.EditorMode
	addModeEngineChan    chan editorEngine.EditorAddMode
	drawModeEngineChan   chan editorEngine.EditorDrawMode
	ControlChannel       chan int

	// Cursor updates can only be applied on the main thread
	cursorUpdate  bool
	currentCursor CustomCursorType
}

func NewCustomCursors(
	globalEditEngineRegChan chan chan editorEngine.EditorMode,
	addModeEngineRegChan chan chan editorEngine.EditorAddMode,
	drawModeEngineRegChan chan chan editorEngine.EditorDrawMode) *CustomCursors {

	cursors := CustomCursors{
		cursors:              make(map[CustomCursorType]*glfw.Cursor),
		drawModeCursors:      make(map[editorEngine.EditorDrawMode]CustomCursorType),
		addModeCursors:       make(map[editorEngine.EditorAddMode]CustomCursorType),
		globalEditEngineChan: make(chan editorEngine.EditorMode, 2),
		addModeEngineChan:    make(chan editorEngine.EditorAddMode, 2),
		drawModeEngineChan:   make(chan editorEngine.EditorDrawMode, 2),
		ControlChannel:       make(chan int),
		cursorUpdate:         true,
		currentCursor:        Selection}

	cursors.loadCursors()

	globalEditEngineRegChan <- cursors.globalEditEngineChan
	addModeEngineRegChan <- cursors.addModeEngineChan
	drawModeEngineRegChan <- cursors.drawModeEngineChan

	go cursors.run()

	return &cursors
}

func (c *CustomCursors) run() {
	for {
		select {
		case newMode := <-c.globalEditEngineChan:
			if newMode == editorEngine.Select {
				c.currentCursor = Selection
				c.cursorUpdate = true
			}
			break
		case drawMode := <-c.drawModeEngineChan:
			c.currentCursor = c.drawModeCursors[drawMode]
			c.cursorUpdate = true
			break
		case addMode := <-c.addModeEngineChan:
			c.currentCursor = c.addModeCursors[addMode]
			c.cursorUpdate = true
			break
		case _ = <-c.ControlChannel:
			return
		}
	}
}

func (c *CustomCursors) loadCursors() {
	c.cursors[Selection] = glfw.CreateStandardCursor(glfw.ArrowCursor)

	// Load all additional cursors
	type CursorPair struct {
		location   string
		cursorType CustomCursorType
	}

	cursorsToLoad := []CursorPair{
		CursorPair{"data/cursors/PowerPlant.png", PowerPlantAdd},
		CursorPair{"data/cursors/PowerLine.png", PowerLineAdd},
		CursorPair{"data/cursors/RoadLine.png", RoadLineAdd},
		CursorPair{"data/cursors/draw/TerrainFlatten.png", TerrainFlatten},
		CursorPair{"data/cursors/draw/TerrainSharpen.png", TerrainSharpen},
		CursorPair{"data/cursors/draw/TerrainTrees.png", TerrainTrees},
		CursorPair{"data/cursors/draw/TerrainShrubs.png", TerrainShrubs},
		CursorPair{"data/cursors/draw/TerrainHills.png", TerrainHills},
		CursorPair{"data/cursors/draw/TerrainValleys.png", TerrainValleys}}

	for _, cursor := range cursorsToLoad {
		cursorImage := commonIo.ReadImageFromFile(cursor.location)
		c.cursors[cursor.cursorType] = glfw.CreateCursor(cursorImage, 0, 0)
	}

	// Assign the maps to simplify routing logic
	c.addModeCursors[editorEngine.PowerLine] = PowerLineAdd
	c.addModeCursors[editorEngine.PowerPlant] = PowerPlantAdd
	c.addModeCursors[editorEngine.RoadLine] = RoadLineAdd
	c.drawModeCursors[editorEngine.TerrainFlatten] = TerrainFlatten
	c.drawModeCursors[editorEngine.TerrainSharpen] = TerrainSharpen
	c.drawModeCursors[editorEngine.TerrainTrees] = TerrainTrees
	c.drawModeCursors[editorEngine.TerrainShrubs] = TerrainShrubs
	c.drawModeCursors[editorEngine.TerrainHills] = TerrainHills
	c.drawModeCursors[editorEngine.TerrainValleys] = TerrainValleys
}

func (c *CustomCursors) Update(window *glfw.Window) {
	if c.cursorUpdate {
		window.SetCursor(c.cursors[c.currentCursor])
	}
}

func (c *CustomCursors) Delete() {
	for _, cursor := range c.cursors {
		cursor.Destroy()
	}
}
