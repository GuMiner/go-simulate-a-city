package gamegrid

import (
	"common/commonmath"
	"common/commonopengl"
	"sim/config"

	"github.com/go-gl/mathgl/mgl32"
)

// Contains computations for visiblity of the graming grid

func getMinMaxVisibleRange(offset mgl32.Vec2, scale float32) (minTile mgl32.Vec2, maxTile mgl32.Vec2) {
	regionSize := config.Config.Terrain.RegionSize

	minTile = MapToBoard(mgl32.Vec2{0, 0}, offset, scale).Mul(1.0 / float32(regionSize))
	maxTile = MapToBoard(mgl32.Vec2{1, 1}, offset, scale).Mul(1.0 / float32(regionSize))
	return minTile, maxTile
}

func ComputeVisibleRegions(offset mgl32.Vec2, scale float32) []commonMath.IntVec2 {
	minTile, maxTile := getMinMaxVisibleRange(offset, scale)

	visibleTiles := make([]commonMath.IntVec2, 0)
	for i := int(minTile.X() - 1.0); i <= int(maxTile.X()+1.0); i++ {
		for j := int(minTile.Y() - 1.0); j <= int(maxTile.Y()+1.0); j++ {
			visibleTiles = append(visibleTiles, commonMath.IntVec2{i, j})
		}
	}

	return visibleTiles
}

func ComputePrecacheRegions(offset mgl32.Vec2, scale float32) []commonMath.IntVec2 {
	minTile, maxTile := getMinMaxVisibleRange(offset, scale)

	xMin := int(minTile.X() - 2.0)
	xMax := int(maxTile.X() + 2.0)
	yMin := int(minTile.Y() - 2.0)
	yMax := int(maxTile.Y() + 2.0)

	visibleTiles := make([]commonMath.IntVec2, 0)
	for i := xMin; i <= xMax; i++ {
		visibleTiles = append(visibleTiles, commonMath.IntVec2{i, yMin})
		visibleTiles = append(visibleTiles, commonMath.IntVec2{i, yMax})
	}

	// Offset by 1 to avoid double-adding corners
	for j := yMin + 1; j <= yMax-1; j++ {
		visibleTiles = append(visibleTiles, commonMath.IntVec2{xMin, j})
		visibleTiles = append(visibleTiles, commonMath.IntVec2{xMax, j})
	}

	return visibleTiles
}

// Maps a (0, 0) to (1, 1) screen position to a board location.
func MapToBoard(screenPos mgl32.Vec2, offset mgl32.Vec2, scale float32) mgl32.Vec2 {
	windowSize := commonOpenGl.GetWindowSize()

	modifiedRegionPos := mgl32.Vec2{(screenPos.X() - 0.5) * windowSize.X(), (screenPos.Y() - 0.5) * windowSize.Y()}
	regionPos := modifiedRegionPos.Mul(1.0 / scale).Add(offset)

	return regionPos
}

// Resizes a full-size region tile to the appropriate scale given the current screen size and zoom factor
// Returns the screen size (a full size tile will span from (0, 0) to (1, 1))
func GetRegionTileScale(zoomFactor float32) mgl32.Vec2 {
	regionSize := config.Config.Terrain.RegionSize
	windowSize := commonOpenGl.GetWindowSize()
	return mgl32.Vec2{
		zoomFactor * float32(regionSize) / windowSize.X(),
		zoomFactor * float32(regionSize) / windowSize.Y()}
}

// Returns the screen position ((0, 0) to (1, 1)) of the region tile requested
func GetRegionTileOffset(x, y int, offset mgl32.Vec2, zoomFactor float32) mgl32.Vec2 {
	regionSize := config.Config.Terrain.RegionSize
	windowSize := commonOpenGl.GetWindowSize()

	regionStart := mgl32.Vec2{float32(x * regionSize), float32(y * regionSize)}
	modifiedRegionStart := regionStart.Sub(offset).Mul(zoomFactor)

	return mgl32.Vec2{modifiedRegionStart.X()/windowSize.X() + 0.5, modifiedRegionStart.Y()/windowSize.Y() + 0.5}
}

// Maps a region on the board to a GLSL (-1, -1) to (1, 1) region
func MapEngineRegionToScreen(region *commonMath.Region, zoomFactor float32, offset mgl32.Vec2) *commonMath.Region {
	// The only variables that are updated (for now) are position and scale
	return &commonMath.Region{
		RegionType:  region.RegionType,
		Orientation: region.Orientation,
		Scale:       region.Scale / zoomFactor,
		Position:    MapPositionToScreen(region.Position, zoomFactor, offset)}
}

func MapPositionToScreen(point mgl32.Vec2, zoomFactor float32, offset mgl32.Vec2) mgl32.Vec2 {
	windowSize := commonOpenGl.GetWindowSize()
	point = point.Sub(offset).Mul(zoomFactor)
	point = mgl32.Vec2{2 * point.X() / windowSize.X(), -2 * point.Y() / windowSize.Y()}
	return point
}
