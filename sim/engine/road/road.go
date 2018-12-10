package road

import (
	"fmt"
	"go-simulate-a-city/sim/core/dto/vehicledto"
	"go-simulate-a-city/sim/core/mailroom"
	"go-simulate-a-city/sim/engine/core/dto"
	"go-simulate-a-city/sim/engine/vehicle"

	"github.com/go-gl/mathgl/mgl32"
)

type progressingVehicle struct {
	vehicle *vehicle.Vehicle
	speed   float32
	percent float32
}

type VehicleAddition struct {
	VehicleId        int64
	Vehicle          *vehicle.Vehicle
	SourceTerminusId int64
	Speed            float32
}

type RoadTerminus struct {
	location mgl32.Vec2

	Id                     int64
	LineAddVehicleChannels map[int64]chan VehicleAddition // TODO: use the graph, not a hardcoded value
	AddVehicleChannel      chan VehicleAddition
	// timerUpdateChannel chan dto.Time TODO: perform time-based updates for intersections
	ControlChannel chan int
}

func NewRoadTerminus(location mgl32.Vec2) *RoadTerminus {
	terminus := RoadTerminus{
		location:               location,
		LineAddVehicleChannels: make(map[int64]chan VehicleAddition),
		AddVehicleChannel:      make(chan VehicleAddition, 3),
		ControlChannel:         make(chan int)}

	return &terminus
}

type RoadLine struct {
	capacity int64

	lowToHighTraffic map[int64]*progressingVehicle
	highToLowTraffic map[int64]*progressingVehicle

	lowTerminus           int64
	lowTerminusAddChannel chan VehicleAddition

	highTerminus           int64
	highTerminusAddChannel chan VehicleAddition

	Id                 int64
	TimerUpdateChannel chan dto.Time
	AddVehicleChannel  chan VehicleAddition
	ControlChannel     chan int
}

func NewRoadLine(capacity int64) *RoadLine {
	roadLine := RoadLine{
		capacity:           capacity,
		lowToHighTraffic:   make(map[int64]*progressingVehicle),
		highToLowTraffic:   make(map[int64]*progressingVehicle),
		TimerUpdateChannel: make(chan dto.Time, 3),
		AddVehicleChannel:  make(chan VehicleAddition, 3),
		ControlChannel:     make(chan int)}

	return &roadLine
}

func (r *RoadLine) run() {
	for {
		select {
		case addition := <-r.AddVehicleChannel:
			if addition.SourceTerminusId == r.lowTerminus {
				r.lowToHighTraffic[addition.VehicleId] = &progressingVehicle{
					vehicle: addition.Vehicle,
					speed:   addition.Speed,
					percent: 0.0}

				mailroom.VehicleUpdateChannel <- vehicledto.VehicleUpdate{
					Id:            addition.VehicleId,
					RoadId:        r.Id,
					TravelLength:  0.001,
					VehicleLength: addition.Vehicle.Length}

			} else {
				r.highToLowTraffic[addition.VehicleId] = &progressingVehicle{
					vehicle: addition.Vehicle,
					speed:   addition.Speed,
					percent: 0.0}

				mailroom.VehicleUpdateChannel <- vehicledto.VehicleUpdate{
					Id:            addition.VehicleId,
					RoadId:        r.Id,
					TravelLength:  -0.001,
					VehicleLength: addition.Vehicle.Length}
			}
		case _ = <-r.TimerUpdateChannel:
			// Move traffic along the road line
			// TODO: Silly demo
			for vehicleId, vehicle := range r.highToLowTraffic {
				fmt.Printf("vehicle %v at L-H %v on line %v from %v to %v\n", vehicleId, vehicle.percent, r.Id, r.highTerminus, r.lowTerminus)
				vehicle.percent += 0.05
				if vehicle.percent >= 1.0 {
					r.lowTerminusAddChannel <- VehicleAddition{
						VehicleId:        vehicleId,
						Vehicle:          vehicle.vehicle,
						SourceTerminusId: r.highTerminus,
						Speed:            vehicle.speed}
					delete(r.highToLowTraffic, vehicleId)
				} else {
					mailroom.VehicleUpdateChannel <- vehicledto.VehicleUpdate{
						Id:            vehicleId,
						RoadId:        r.Id,
						TravelLength:  -vehicle.percent,
						VehicleLength: vehicle.vehicle.Length}
				}
			}

			for vehicleId, vehicle := range r.lowToHighTraffic {
				fmt.Printf("vehicle %v at H-L %v on line %v from %v to %v\n", vehicleId, vehicle.percent, r.Id, r.lowTerminus, r.highTerminus)
				vehicle.percent += 0.05
				if vehicle.percent >= 1.0 {
					r.highTerminusAddChannel <- VehicleAddition{
						VehicleId:        vehicleId,
						Vehicle:          vehicle.vehicle,
						SourceTerminusId: r.lowTerminus,
						Speed:            vehicle.speed}
					delete(r.lowToHighTraffic, vehicleId)
				} else {
					mailroom.VehicleUpdateChannel <- vehicledto.VehicleUpdate{
						Id:            vehicleId,
						RoadId:        r.Id,
						TravelLength:  vehicle.percent,
						VehicleLength: vehicle.vehicle.Length}
				}
			}
		case _ = <-r.ControlChannel:
			return
		}
	}
}

func (r *RoadTerminus) run() {
	for {
		select {
		case vehicle := <-r.AddVehicleChannel:
			for destinationId, _ := range r.LineAddVehicleChannels {
				fmt.Printf("(%v) %v -- %v\n", r.Id, vehicle.SourceTerminusId, destinationId)
			}

			// TODO silly demo logic.
			moved := false
			for destinationId, channel := range r.LineAddVehicleChannels {
				if destinationId != vehicle.SourceTerminusId {
					// We're going somewhere else, so send it!
					channel <- VehicleAddition{
						VehicleId:        vehicle.VehicleId,
						Vehicle:          vehicle.Vehicle,
						Speed:            vehicle.Speed,
						SourceTerminusId: r.Id}
					moved = true
					break
				}
			}

			if !moved {
				// The vehicle has no where else to go so it bounces to the first result
				for _, channel := range r.LineAddVehicleChannels {
					channel <- VehicleAddition{
						VehicleId:        vehicle.VehicleId,
						Vehicle:          vehicle.Vehicle,
						Speed:            vehicle.Speed,
						SourceTerminusId: r.Id}
					break
				}
			}

			// Move vehicle through the intersection, or
			// to the next line for disjointed segments
		case _ = <-r.ControlChannel:
			return
		}
	}
}
