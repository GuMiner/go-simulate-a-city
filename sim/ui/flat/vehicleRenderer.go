package flat

import (
	"fmt"
	"go-simulate-a-city/sim/core/dto/geometry"
	"go-simulate-a-city/sim/core/dto/vehicledto"

	"github.com/go-gl/mathgl/mgl32"
)

type VehicleRenderer struct {
	roadTerminii map[int64]mgl32.Vec2
	roadLines    map[int64]geometry.IdOnlyLine

	VehicleUpdateChannel   chan vehicledto.VehicleUpdate
	VehicleDeletionChannel chan int64
	RoadLineRegChannel     chan geometry.IdOnlyLine
	TerminusChannel        chan geometry.IdPoint
	Renderer               *LineRenderer
}

func NewVehicleRenderer() *VehicleRenderer {
	renderer := VehicleRenderer{
		roadTerminii:           make(map[int64]mgl32.Vec2),
		roadLines:              make(map[int64]geometry.IdOnlyLine),
		VehicleUpdateChannel:   make(chan vehicledto.VehicleUpdate, 3),
		VehicleDeletionChannel: make(chan int64, 3),
		RoadLineRegChannel:     make(chan geometry.IdOnlyLine, 3),
		TerminusChannel:        make(chan geometry.IdPoint, 3),
		Renderer:               NewLineRenderer(mgl32.Vec3{1, 1, 0})}

	go renderer.run()
	return &renderer
}

func (r *VehicleRenderer) run() {
	for {
		select {
		case roadLine := <-r.RoadLineRegChannel:
			r.roadLines[roadLine.Id] = roadLine
		case terminus := <-r.TerminusChannel:
			r.roadTerminii[terminus.Id] = terminus.Point
		case vehicleUpdate := <-r.VehicleUpdateChannel:
			if road, ok := r.roadLines[vehicleUpdate.RoadId]; ok {
				if startPos, ok := r.roadTerminii[road.Start]; ok {
					if endPos, ok := r.roadTerminii[road.End]; ok {
						fmt.Printf("Valid vehicle update: %v %v %v\n", startPos, endPos, road)
						// Swap so the percentage we take is always from start to end.
						if (road.End < road.Start && vehicleUpdate.TravelLength > 0) ||
							(road.End > road.Start && vehicleUpdate.TravelLength < 0) {
							temp := startPos
							startPos = endPos
							endPos = temp
						}

						if vehicleUpdate.TravelLength < 0 {
							vehicleUpdate.TravelLength = -vehicleUpdate.TravelLength
						}

						// Compute start and end
						roadSegment := endPos.Sub(startPos)
						vehicleLengthPercent := vehicleUpdate.VehicleLength / roadSegment.Len()
						start := roadSegment.Mul(vehicleUpdate.TravelLength).Add(startPos)
						end := roadSegment.Mul(vehicleUpdate.TravelLength + vehicleLengthPercent).Add(startPos)

						r.Renderer.NewLineChannel <- geometry.NewIdLine(vehicleUpdate.Id, [2]mgl32.Vec2{start, end})
					}
				}
			}
		case vehicleId := <-r.VehicleDeletionChannel:
			r.Renderer.DeleteLineChannel <- vehicleId
		}
	}
}
