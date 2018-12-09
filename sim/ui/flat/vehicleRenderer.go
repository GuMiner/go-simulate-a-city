package flat

import (
	"github.com/go-gl/mathgl/mgl32"
)

// TODO: DTO
type VehicleUpdate struct {
	Id            int64
	RoadId        int64
	TravelLength  float32 // Negative -- high to low. Positive -- low to high.
	VehicleLength float32
}

type RoadLineRegistration struct {
	Id            int64
	StartTerminus int64
	EndTerminus   int64
}

type RoadTerminus struct {
	Id       int64
	Position mgl32.Vec2
}

type VehicleRenderer struct {
	roadTerminii map[int64]mgl32.Vec2
	roadLines    map[int64]RoadLineRegistration

	VehicleUpdateChannel chan VehicleUpdate
	RoadLineRegChannel   chan RoadLineRegistration
	TerminusChannel      chan RoadTerminus
	Renderer             *LineRenderer
}

func NewVehicleRenderer() *VehicleRenderer {
	renderer := VehicleRenderer{
		Renderer: NewLineRenderer(mgl32.Vec3{1, 1, 0})}

	go renderer.run()
	return &renderer
}

func (r *VehicleRenderer) run() {
	for {

	}
}
