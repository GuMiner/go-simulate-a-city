package terrain

import (
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/engine/subtile"

	"github.com/go-gl/mathgl/mgl32"
)

type TerrainMap struct {
	registeredNewTerrainChannels []chan *TerrainUpdate

	SubMaps              map[int]map[int]*TerrainSubMap
	NewTerrainRegChannel chan chan *TerrainUpdate
	ControlChannel       chan int
}

func NewTerrainMap() *TerrainMap {
	terrainMap := TerrainMap{
		registeredNewTerrainChannels: make([]chan *TerrainUpdate, 0),
		NewTerrainRegChannel:         make(chan chan *TerrainUpdate),
		ControlChannel:               make(chan int),
		SubMaps:                      make(map[int]map[int]*TerrainSubMap)}

	go terrainMap.run()

	return &terrainMap
}

func (t *TerrainMap) run() {
	for {
		select {
		case reg := <-t.NewTerrainRegChannel:
			t.registeredNewTerrainChannels = append(t.registeredNewTerrainChannels, reg)
			break
		case _ = <-t.ControlChannel:
			return
		}
	}
}

func (t *TerrainMap) GetOrAddRegion(x, y int) *TerrainSubMap {
	if _, ok := t.SubMaps[x]; !ok {
		t.SubMaps[x] = make(map[int]*TerrainSubMap)
	}

	if _, ok := t.SubMaps[x][y]; !ok {
		t.SubMaps[x][y] = NewTerrainSubMap(x, y)
		for _, reg := range t.registeredNewTerrainChannels {
			reg <- NewTerrainUpdate(t.SubMaps[x][y], x, y)
		}
	}

	return t.SubMaps[x][y]
}

func (t *TerrainMap) ValidateGroundLocation(reg commonMath.Region) bool {

	iterate := func(x, y int) bool {
		pos := mgl32.Vec2{float32(x), float32(y)}
		texel, _ := t.getTexel(pos)

		return texel.TerrainType == Water
	}

	return !reg.IterateIntWithEarlyExit(iterate)
}

func (t *TerrainMap) Flatten(region commonMath.Region, amount float32) {
	t.performRegionBasedUpdate(region, amount, flatten)
}

func (t *TerrainMap) Sharpen(region commonMath.Region, amount float32) {
	t.performRegionBasedUpdate(region, amount, sharpen)
}

func (t *TerrainMap) Hills(region commonMath.Region, amount float32) {
	t.performRegionBasedUpdate(region, amount, hills)
}

func (t *TerrainMap) Valleys(region commonMath.Region, amount float32) {
	t.performRegionBasedUpdate(region, amount, valleys)
}

func (t *TerrainMap) performRegionBasedUpdate(region commonMath.Region, amount float32, update func(mgl32.Vec2, mgl32.Vec2, *TerrainTexel, float32, float32, float32)) {
	centerTexel, _ := t.getTexel(region.Position)
	centralHeight := centerTexel.Height

	region.IterateIntWithEarlyExit(func(x, y int) bool {
		modifiedPos := mgl32.Vec2{float32(x) + 0.5, float32(y) + 0.5}
		texel, texelRegion := t.getTexel(modifiedPos)

		update(region.Position, modifiedPos, texel, centralHeight, amount, region.Scale)
		for _, reg := range t.registeredNewTerrainChannels {
			reg <- NewTerrainUpdate(texelRegion, x, y)
		}

		// Never early exit
		return false
	})
}

// Average, moving parts that are farther away closer in faster.
func flatten(centerPosition, texelPosition mgl32.Vec2, texel *TerrainTexel, centerHeight, amount, regionSize float32) {
	heightDifference := texel.Height - centerHeight
	texel.Height = texel.Height - heightDifference*amount
	texel.Normalize()
}

// Reverse average, moving parts that are farther away further faster.
func sharpen(centerPosition, texelPosition mgl32.Vec2, texel *TerrainTexel, centerHeight, amount, regionSize float32) {
	heightDifference := texel.Height - centerHeight
	texel.Height = texel.Height + heightDifference*amount
	texel.Normalize()
}

// Makes hills, pushing pixels near the center position upwards,
func hills(centerPosition, texelPosition mgl32.Vec2, texel *TerrainTexel, centerHeight, amount, regionSize float32) {
	distanceFactor := 1.0 - centerPosition.Sub(texelPosition).Len()/regionSize

	texel.Height = texel.Height + amount*distanceFactor
	texel.Normalize()
}

func valleys(centerPosition, texelPosition mgl32.Vec2, texel *TerrainTexel, centerHeight, amount, regionSize float32) {
	distanceFactor := 1.0 - centerPosition.Sub(texelPosition).Len()/regionSize

	texel.Height = texel.Height - amount*distanceFactor
	texel.Normalize()
}

func (t *TerrainMap) getTexel(pos mgl32.Vec2) (*TerrainTexel, *TerrainSubMap) {
	regionX, regionY := subtile.GetRegionIndices(pos, config.Config.Terrain.RegionSize)
	region := t.GetOrAddRegion(regionX, regionY)

	localX, localY := subtile.GetLocalIndices(pos, regionX, regionY, config.Config.Terrain.RegionSize)
	return &region.Texels[localX][localY], region
}
