package mailroom

import (
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/core/dto/editorengdto"
	"go-simulate-a-city/sim/core/dto/terraindto"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

// Defines the mailroom designed to easily connect the core channels making up the game engine
// This does make it slightly harder to ensure proper first-time setup, but significantly simplifies doing that setup.

// Generic input
var MousePressedRegChannel chan chan glfw.MouseButton
var MouseReleasedRegChannel chan chan glfw.MouseButton

// Camera
var CameraOffsetRegChannel chan chan mgl32.Vec2
var CameraScaleRegChannel chan chan float32
var BoardPosChangeRegChannel chan chan mgl32.Vec2

// Terrain
var NewTerrainRegChannel chan chan *terraindto.TerrainUpdate
var NewRegionRegChannel chan chan commonMath.IntVec2

// Editor engine
var EngineModeRegChannel chan chan editorengdto.EditorMode
var EngineAddModeRegChannel chan chan editorengdto.EditorAddMode
var EngineDrawModeRegChannel chan chan editorengdto.EditorDrawMode

// Power Plants
var NewPlantRegionRegChannel chan chan *commonMath.Region
