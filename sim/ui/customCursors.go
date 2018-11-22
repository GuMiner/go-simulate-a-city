package ui

import (
	"go-simulate-a-city/common/commonio"
	"go-simulate-a-city/sim/core/dto/editorengdto"
	"go-simulate-a-city/sim/core/mailroom"

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
	drawModeCursors map[editorengdto.EditorDrawMode]CustomCursorType
	addModeCursors  map[editorengdto.EditorAddMode]CustomCursorType

	globalEditEngineChan chan editorengdto.EditorMode
	addModeEngineChan    chan editorengdto.EditorAddMode
	drawModeEngineChan   chan editorengdto.EditorDrawMode
	ControlChannel       chan int

	// Cursor updates can only be applied on the main thread
	cursorUpdate  bool
	currentCursor CustomCursorType
}

func NewCustomCursors() *CustomCursors {

	cursors := CustomCursors{
		cursors:              make(map[CustomCursorType]*glfw.Cursor),
		drawModeCursors:      make(map[editorengdto.EditorDrawMode]CustomCursorType),
		addModeCursors:       make(map[editorengdto.EditorAddMode]CustomCursorType),
		globalEditEngineChan: make(chan editorengdto.EditorMode, 2),
		addModeEngineChan:    make(chan editorengdto.EditorAddMode, 2),
		drawModeEngineChan:   make(chan editorengdto.EditorDrawMode, 2),
		ControlChannel:       make(chan int),
		cursorUpdate:         true,
		currentCursor:        Selection}

	cursors.loadCursors()

	mailroom.EngineModeRegChannel <- cursors.globalEditEngineChan
	mailroom.EngineAddModeRegChannel <- cursors.addModeEngineChan
	mailroom.EngineDrawModeRegChannel <- cursors.drawModeEngineChan

	go cursors.run()

	return &cursors
}

func (c *CustomCursors) run() {
	for {
		select {
		case newMode := <-c.globalEditEngineChan:
			if newMode == editorengdto.Select {
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
	c.addModeCursors[editorengdto.PowerLine] = PowerLineAdd
	c.addModeCursors[editorengdto.PowerPlant] = PowerPlantAdd
	c.addModeCursors[editorengdto.RoadLine] = RoadLineAdd
	c.drawModeCursors[editorengdto.TerrainFlatten] = TerrainFlatten
	c.drawModeCursors[editorengdto.TerrainSharpen] = TerrainSharpen
	c.drawModeCursors[editorengdto.TerrainTrees] = TerrainTrees
	c.drawModeCursors[editorengdto.TerrainShrubs] = TerrainShrubs
	c.drawModeCursors[editorengdto.TerrainHills] = TerrainHills
	c.drawModeCursors[editorengdto.TerrainValleys] = TerrainValleys
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
