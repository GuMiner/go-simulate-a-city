package terrain

import (
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/core/gamegrid"
	"go-simulate-a-city/sim/engine/subtile"

	"github.com/go-gl/mathgl/mgl32"
)

var FORCE_REFRESH int = 2

type TerrainMap struct {
	cameraOffset                 mgl32.Vec2
	cameraScale                  float32
	offsetChangeChannel          chan mgl32.Vec2
	scaleChangeChannel           chan float32
	registeredNewTerrainChannels []chan *TerrainUpdate
	registeredNewRegionChannels  []chan commonMath.IntVec2

	SubMaps              map[int]map[int]*TerrainSubMap
	NewTerrainRegChannel chan chan *TerrainUpdate
	NewRegionRegChannel  chan chan commonMath.IntVec2
	ControlChannel       chan int
}

func NewTerrainMap(
	camOffsetRegChannel chan chan mgl32.Vec2,
	camScaleRegChannel chan chan float32) *TerrainMap {
	terrainMap := TerrainMap{
		cameraOffset:                 mgl32.Vec2{0, 0},
		cameraScale:                  1.0,
		offsetChangeChannel:          make(chan mgl32.Vec2, 3),
		scaleChangeChannel:           make(chan float32, 3),
		registeredNewTerrainChannels: make([]chan *TerrainUpdate, 0),
		registeredNewRegionChannels:  make([]chan commonMath.IntVec2, 0),
		NewTerrainRegChannel:         make(chan chan *TerrainUpdate),
		NewRegionRegChannel:          make(chan chan commonMath.IntVec2),
		ControlChannel:               make(chan int),
		SubMaps:                      make(map[int]map[int]*TerrainSubMap)}

	camOffsetRegChannel <- terrainMap.offsetChangeChannel
	camScaleRegChannel <- terrainMap.scaleChangeChannel

	go terrainMap.run()

	return &terrainMap
}

func (t *TerrainMap) precacheRegions() {
	regions := gamegrid.ComputePrecacheRegions(t.cameraOffset, t.cameraScale)
	for _, region := range regions {
		_ = t.GetOrAddRegion(region.X(), region.Y())
	}
}

func (t *TerrainMap) run() {
	for {
		select {
		case t.cameraOffset = <-t.offsetChangeChannel:
			t.precacheRegions()
			break
		case t.cameraScale = <-t.scaleChangeChannel:
			t.precacheRegions()
			break
		case reg := <-t.NewTerrainRegChannel:
			t.registeredNewTerrainChannels = append(t.registeredNewTerrainChannels, reg)
			break
		case reg := <-t.NewRegionRegChannel:
			t.registeredNewRegionChannels = append(t.registeredNewRegionChannels, reg)
		case control := <-t.ControlChannel:
			if control == FORCE_REFRESH {
				regions := gamegrid.ComputeVisibleRegions(t.cameraOffset, t.cameraScale)
				for _, region := range regions {
					_ = t.GetOrAddRegion(region.X(), region.Y())
				}
			} else {
				return
			}
		}
	}
}

func (t *TerrainMap) GetOrAddRegion(x, y int) *TerrainSubMap {
	if _, ok := t.SubMaps[x]; !ok {
		t.SubMaps[x] = make(map[int]*TerrainSubMap)
	}

	if _, ok := t.SubMaps[x][y]; !ok {
		t.SubMaps[x][y] = NewTerrainSubMap(x, y)
		terrainUpdate := NewTerrainUpdate(t.SubMaps[x][y], x, y)
		for _, reg := range t.registeredNewTerrainChannels {
			reg <- terrainUpdate
		}

		for _, reg := range t.registeredNewRegionChannels {
			reg <- commonMath.IntVec2{x, y}
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
		terrainUpdate := NewTerrainUpdate(texelRegion, x, y)
		for _, reg := range t.registeredNewTerrainChannels {
			reg <- terrainUpdate
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
