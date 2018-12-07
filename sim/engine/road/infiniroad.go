package road

import (
	"fmt"
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/core/mailroom"

	"github.com/ojrac/opensimplex-go"

	"github.com/go-gl/mathgl/mgl32"
)

// Defines the node ends of the infinitely generated road.
// These may become invalid as roads are deleted.
type InfiniRoadNodeEnds struct {
	// 0 == West, 1 == East. Simplifies the math below
	RoadEnds [2]int64
}

type InfiniRoadGenerator struct {
	grid *RoadGrid

	noise            *opensimplex.Noise
	newRegionChannel chan commonMath.IntVec2

	// Defines if each automatically-generated road has been generated
	RoadGenerated map[int]map[int]bool
	RoadNodeEdges map[int]map[int]InfiniRoadNodeEnds
}

func NewInfiniRoadGenerator(grid *RoadGrid) *InfiniRoadGenerator {
	infiniRoadGenerator := InfiniRoadGenerator{
		grid:             grid,
		noise:            opensimplex.NewWithSeed(int64(42)), // TODO: Configurable??
		newRegionChannel: make(chan commonMath.IntVec2, 3),
		RoadGenerated:    make(map[int]map[int]bool),
		RoadNodeEdges:    make(map[int]map[int]InfiniRoadNodeEnds)}

	mailroom.NewRegionRegChannel <- infiniRoadGenerator.newRegionChannel

	go infiniRoadGenerator.run()
	return &infiniRoadGenerator
}

func (i *InfiniRoadGenerator) run() {
	for {
		select {
		case newRegion := <-i.newRegionChannel:
			i.GenerateRoad(newRegion)
			break
		}
	}
}

func (i *InfiniRoadGenerator) markRoadAsGenerated(region commonMath.IntVec2) {
	i.RoadGenerated[region.X()][region.Y()] = true
}

func (i *InfiniRoadGenerator) addMapSliceIfMissing(x int) {
	if _, ok := i.RoadGenerated[x]; !ok {
		i.RoadGenerated[x] = make(map[int]bool)
	}
}

func (i *InfiniRoadGenerator) getNodeId(region commonMath.IntVec2, offsetX int) int64 {
	effectiveRegion := commonMath.IntVec2{region.X() + offsetX, region.Y()}
	roadEndIndex := -offsetX
	roadEndIndex = commonMath.MaxInt(0, roadEndIndex)

	i.addMapSliceIfMissing(effectiveRegion.X())
	if i.RoadGenerated[effectiveRegion.X()][effectiveRegion.Y()] {
		return i.RoadNodeEdges[effectiveRegion.X()][effectiveRegion.Y()].RoadEnds[roadEndIndex]
	}

	return -1
}

func (i *InfiniRoadGenerator) GenerateRoad(region commonMath.IntVec2) {
	i.addMapSliceIfMissing(region.X())
	if region.Y() != 0 {
		i.markRoadAsGenerated(region)
		return
	}

	westNodeId := i.getNodeId(region, -1)
	eastNodeId := i.getNodeId(region, 1)
	fmt.Printf("%v, %v, %v, %v\n", region.X(), region.Y(), westNodeId, eastNodeId)

	fRegionSize := float32(config.Config.Terrain.RegionSize)
	halfYHeight := fRegionSize / 2

	// Throw in a bit of an offset so straight lines have noticeable nodes for connection
	scale := 30.0
	startOffset := i.noise.Eval2(float64(region.X()), float64(region.Y())) * scale
	endOffset := i.noise.Eval2(float64(region.X())+0.5, float64(region.Y())+0.5) * scale

	start := mgl32.Vec2{float32(region.X()) * fRegionSize, float32(region.Y())*halfYHeight + float32(startOffset)}
	end := mgl32.Vec2{float32(region.X()+1) * fRegionSize, float32(region.Y())*halfYHeight + float32(endOffset)}

	// Validate the nodes still exist if indicated. If they do, update the positions
	// If they don't reset this so we don't attempt to connect to non-existing nodes
	if westNodeId != -1 {
		if roadTerminus := i.grid.grid.GetNode(westNodeId); roadTerminus != nil {
			start = roadTerminus.(*RoadTerminus).location
		} else {
			westNodeId = -1
		}
	}

	if eastNodeId != -1 {
		if roadTerminus := i.grid.grid.GetNode(eastNodeId); roadTerminus != nil {
			end = roadTerminus.(*RoadTerminus).location
		} else {
			eastNodeId = -1
		}
	}

	// TODO: Default to highway capacity for the infinte road.
	// TODO: This should be a lot smarter and follow contours
	roadId := int64(-1)
	westNodeId, roadId, eastNodeId = i.grid.AddLine(start, end, 1000, westNodeId, eastNodeId)
	fmt.Printf("  Generated new infinite-road element for [%v, %v]: %v\n", region.X(), region.Y(), roadId)

	// Update our caches so we don't infinitely generate infinite roads.
	i.markRoadAsGenerated(region)
	if _, ok := i.RoadNodeEdges[region.X()]; !ok {
		i.RoadNodeEdges[region.X()] = make(map[int]InfiniRoadNodeEnds)
	}

	i.RoadNodeEdges[region.X()][region.Y()] = InfiniRoadNodeEnds{
		RoadEnds: [2]int64{westNodeId, eastNodeId}}
}
