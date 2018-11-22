package mailroom

import (
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/core/dto/terraindto"

	"github.com/go-gl/mathgl/mgl32"
)

// Defines the mailroom designed to easily connect the core channels making up the game engine

// Camera
var CameraOffsetRegChannel chan chan mgl32.Vec2
var CameraScaleRegChannel chan chan float32
var BoardPosChangeRegChannel chan chan mgl32.Vec2

// Terrain
var NewTerrainRegChannel chan chan *terraindto.TerrainUpdate
var NewRegionRegChannel chan chan commonMath.IntVec2
