package mailroom

import (
	"common/commonmath"
	"sim/core/dto/editorengdto"
	"sim/core/dto/geometry"
	"sim/core/dto/terraindto"
	"sim/core/dto/vehicledto"
	"sim/engine/core/dto"

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
var SnapSettingsRegChannel chan chan editorengdto.SnapSetting
var EngineCancelChannel chan chan bool

// Engine temporal updates
var CoreTimerRegChannel chan chan dto.Time

// --- Rendering ---
// Power
var NewPowerLineChannel chan geometry.IdLine
var DeletePowerLineChannel chan int64

var NewPowerPlantChannel chan geometry.IdRegion
var DeletePowerPlantChannel chan int64

// Road Lines
var NewRoadLineChannel chan geometry.IdLine
var DeleteRoadLineChannel chan int64

// Vehicles
var VehicleUpdateChannel chan vehicledto.VehicleUpdate
var NewRoadLineIdChannel chan geometry.IdOnlyLine
var NewRoadTerminusChannel chan geometry.IdPoint

// Snap nodes
var SnappedNodesUpdateChannel chan []mgl32.Vec2

func Init() {
	// Future implementation
}
