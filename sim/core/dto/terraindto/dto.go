package terraindto

import (
	"common/commonmath"
	"sim/config"
)

type GenerationFunc func(int, int, int, int) []float32

type TerrainTexel struct {
	TerrainType TerrainType

	// Absolute height
	Height float32

	// Relative height for the given terrain type
	HeightPercent float32
}

func (t *TerrainTexel) Normalize() {
	t.Height = commonMath.MinFloat32(1, commonMath.MaxFloat32(0, t.Height))
	t.TerrainType, t.HeightPercent = GetTerrainType(t.Height)
}

type TerrainSubMap struct {
	Texels [][]TerrainTexel
}

func NewTerrainSubMap(x, y int, generationFunction GenerationFunc) *TerrainSubMap {
	regionSize := config.Config.Terrain.RegionSize

	terrainSubMap := TerrainSubMap{
		Texels: make([][]TerrainTexel, regionSize*regionSize)}

	for i := 0; i < regionSize; i++ {
		terrainSubMap.Texels[i] = make([]TerrainTexel, regionSize)
	}

	terrainSubMap.GenerateSubMap(x, y, generationFunction)

	return &terrainSubMap
}

func (t *TerrainSubMap) GenerateSubMap(x, y int, generationFunction GenerationFunc) {
	regionSize := config.Config.Terrain.RegionSize
	heights := generationFunction(regionSize, regionSize, x*regionSize, y*regionSize)
	for i := 0; i < regionSize; i++ {
		for j := 0; j < regionSize; j++ {
			height := heights[i+j*regionSize]
			t.Texels[i][j] = TerrainTexel{Height: height}
			t.Texels[i][j].Normalize()
		}
	}
}

type TerrainUpdate struct {
	Texels [][]TerrainTexel
	Pos    commonMath.IntVec2
}

func NewTerrainUpdate(subMap *TerrainSubMap, x, y int) *TerrainUpdate {
	terrainUpdate := TerrainUpdate{
		Texels: make([][]TerrainTexel, len(subMap.Texels)),
		Pos:    commonMath.IntVec2{x, y}}

	for x := range subMap.Texels {
		terrainUpdate.Texels[x] = make([]TerrainTexel, len(subMap.Texels[x]))
		copy(terrainUpdate.Texels[x], subMap.Texels[x])
	}
	return &terrainUpdate
}
