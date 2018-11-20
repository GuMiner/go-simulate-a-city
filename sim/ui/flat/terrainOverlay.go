package flat

import (
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/core/gamegrid"
	"go-simulate-a-city/sim/engine/terrain"
	"go-simulate-a-city/sim/ui/overlay"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type TerrainOverlay struct {
	textureId uint32
	overlay   *overlay.Overlay
}

func NewTerrainOverlay(textureId uint32) *TerrainOverlay {
	overlay := TerrainOverlay{
		textureId: textureId,
		overlay:   overlay.NewOverlay()}

	return &overlay
}

func (t *TerrainOverlay) GetOverlay() *overlay.Overlay {
	return t.overlay
}

func (t *TerrainOverlay) UpdateCameraOffset(x, y int, offset mgl32.Vec2, zoomFactor float32) {
	regionOffset := gamegrid.GetRegionOffset(x, y, offset, zoomFactor)
	scale := gamegrid.GetRegionScale(zoomFactor)

	t.overlay.UpdateLocation(regionOffset, scale, 1.0)
}

func (t *TerrainOverlay) SetTerrain(texels [][]terrain.TerrainTexel) {
	regionSize := len(texels[0])
	byteTerrain := make([]uint8, regionSize*regionSize*4)
	for i := 0; i < regionSize; i++ {
		for j := 0; j < regionSize; j++ {
			height := texels[i][j].Height

			color, percent := getTerrainColor(height)
			byteTerrain[(i+j*regionSize)*4] = uint8(color.X() * percent)
			byteTerrain[(i+j*regionSize)*4+1] = uint8(color.Y() * percent)
			byteTerrain[(i+j*regionSize)*4+2] = uint8(color.Z() * percent)
			byteTerrain[(i+j*regionSize)*4+3] = 1.0
		}
	}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.textureId)
	gl.TexStorage2D(gl.TEXTURE_2D, 1, gl.RGBA8, int32(regionSize), int32(regionSize))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.TexSubImage2D(gl.TEXTURE_2D, 0,
		0, 0, int32(regionSize), int32(regionSize),
		gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(byteTerrain))

	t.overlay.UpdateTexture(t.textureId)
}

// Given a height, returns the terrain color and percentage within that level
func getTerrainColor(height float32) (mgl32.Vec3, float32) {
	terrainType, percent := terrain.GetTerrainType(height)

	switch terrainType {
	case terrain.Water:
		return config.Config.Ui.TerrainUi.WaterColor.ToVec3(), percent
	case terrain.Sand:
		return config.Config.Ui.TerrainUi.SandColor.ToVec3(), percent
	case terrain.Grass:
		return config.Config.Ui.TerrainUi.GrassColor.ToVec3(), percent
	case terrain.Hills:
		return config.Config.Ui.TerrainUi.HillColor.ToVec3(), percent
	case terrain.Rocks:
		return config.Config.Ui.TerrainUi.RockColor.ToVec3(), percent
	default:
		return config.Config.Ui.TerrainUi.SnowColor.ToVec3(), percent
	}
}
