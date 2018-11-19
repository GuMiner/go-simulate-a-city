package terrain

import (
	"fmt"
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/config"
)

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

func NewTerrainSubMap(x, y int) *TerrainSubMap {
	regionSize := config.Config.Terrain.RegionSize

	terrainSubMap := TerrainSubMap{
		Texels: make([][]TerrainTexel, regionSize*regionSize)}

	for i := 0; i < regionSize; i++ {
		terrainSubMap.Texels[i] = make([]TerrainTexel, regionSize)
	}

	terrainSubMap.GenerateSubMap(x, y)

	return &terrainSubMap
}

func (t *TerrainSubMap) GenerateSubMap(x, y int) {
	regionSize := config.Config.Terrain.RegionSize
	heights := Generate(regionSize, regionSize, x*regionSize, y*regionSize)
	for i := 0; i < regionSize; i++ {
		for j := 0; j < regionSize; j++ {
			height := heights[i+j*regionSize]
			t.Texels[i][j] = TerrainTexel{Height: height}
			t.Texels[i][j].Normalize()
		}
	}

	fmt.Printf("Generated sub map terrain for [%v, %v]\n", x, y)
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
