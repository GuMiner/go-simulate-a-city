package road

import (
	"fmt"
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/core/mailroom"
	"go-simulate-a-city/sim/engine/core/dto"
	"go-simulate-a-city/sim/engine/vehicle"

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
	grid           *RoadGrid
	vehicleManager *vehicle.VehicleManager

	noise              *opensimplex.Noise
	newRegionChannel   chan commonMath.IntVec2
	timerUpdateChannel chan dto.Time

	// Defines if each automatically-generated road has been generated
	RoadGenerated  map[int]bool
	RoadNodeEdges  map[int]InfiniRoadNodeEnds
	WestEdge       int
	EastEdge       int
	WestLineId     int64
	EastLineId     int64
	WestTerminusId int64
	EastTerminusId int64

	NewCarTimer int
}

func NewInfiniRoadGenerator(grid *RoadGrid, vehicleManager *vehicle.VehicleManager) *InfiniRoadGenerator {
	infiniRoadGenerator := InfiniRoadGenerator{
		grid:               grid,
		vehicleManager:     vehicleManager,
		noise:              opensimplex.NewWithSeed(int64(42)), // TODO: Configurable??
		newRegionChannel:   make(chan commonMath.IntVec2, 3),
		timerUpdateChannel: make(chan dto.Time, 3),
		RoadGenerated:      make(map[int]bool),
		RoadNodeEdges:      make(map[int]InfiniRoadNodeEnds),
		WestEdge:           0,
		EastEdge:           0,
		WestLineId:         -1,
		EastLineId:         -1,
		WestTerminusId:     -1,
		EastTerminusId:     -1,
		NewCarTimer:        0}

	mailroom.NewRegionRegChannel <- infiniRoadGenerator.newRegionChannel
	mailroom.CoreTimerRegChannel <- infiniRoadGenerator.timerUpdateChannel

	go infiniRoadGenerator.run()
	return &infiniRoadGenerator
}

func (i *InfiniRoadGenerator) run() {
	for {
		select {
		case newRegion := <-i.newRegionChannel:
			i.GenerateRoad(newRegion)
		case _ = <-i.timerUpdateChannel:
			if i.WestLineId != -1 {
				i.NewCarTimer++

				// This right now is approximately a car each direction each second.
				if i.NewCarTimer == 10 {
					// i.NewCarTimer = 11

					// TODO create cars based on demand and if roads have space
					westVehicle, westVehicleId := i.vehicleManager.NewVehicle()
					fmt.Printf("Adding vehicle %v to %v, line %v\n", westVehicleId, i.WestTerminusId, i.WestLineId)

					// Create a new west-bound car
					westRoadLine := i.grid.grid.GetConnection(i.WestLineId).(*RoadLine)
					westRoadLine.AddVehicleChannel <- VehicleAddition{
						VehicleId:  westVehicleId,
						Vehicle:    westVehicle,
						TerminusId: i.WestTerminusId,
						Speed:      0.0}

					eastVehicle, eastVehicleId := i.vehicleManager.NewVehicle()
					fmt.Printf("Adding vehicle %v to %v, line %v\n", eastVehicleId, i.EastTerminusId, i.EastLineId)

					eastRoadLine := i.grid.grid.GetConnection(i.EastLineId).(*RoadLine)
					eastRoadLine.AddVehicleChannel <- VehicleAddition{
						VehicleId:  eastVehicleId,
						Vehicle:    eastVehicle,
						TerminusId: i.EastTerminusId,
						Speed:      0.0}
				}
			}
		}
	}
}

func (i *InfiniRoadGenerator) markRoadAsGenerated(regionX int) {
	i.RoadGenerated[regionX] = true
}

func (i *InfiniRoadGenerator) getNodeId(regionX, offsetX int) int64 {
	effectiveRegion := regionX + offsetX
	roadEndIndex := -offsetX
	roadEndIndex = commonMath.MaxInt(0, roadEndIndex)

	if i.RoadGenerated[effectiveRegion] {
		return i.RoadNodeEdges[effectiveRegion].RoadEnds[roadEndIndex]
	}

	return -1
}

func (i *InfiniRoadGenerator) GenerateRoad(region commonMath.IntVec2) {
	if region.Y() != 0 {
		return
	}

	fmt.Printf("Max infinite road bounds: %v, %v\n", i.WestEdge, i.EastEdge)

	westNodeId := i.getNodeId(region.X(), -1)
	eastNodeId := i.getNodeId(region.X(), 1)

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

	if region.X()-1 < i.WestEdge {
		i.WestEdge = region.X() - 1
		i.WestLineId = roadId
		i.WestTerminusId = westNodeId
	}

	if region.X()+1 > i.EastEdge {
		i.EastEdge = region.X() + 1
		i.EastLineId = roadId
		i.EastTerminusId = eastNodeId
	}

	// Update our caches so we don't infinitely generate infinite roads.
	i.markRoadAsGenerated(region.X())

	i.RoadNodeEdges[region.X()] = InfiniRoadNodeEnds{RoadEnds: [2]int64{westNodeId, eastNodeId}}
}
