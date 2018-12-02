package flat

import (
	"github.com/go-gl/mathgl/mgl32"
)

type PowerGridRenderer struct {
	LineRenderer  *LineRenderer
	PlantRenderer *RegionRenderer
}

func NewPowerGridRenderer() *PowerGridRenderer {
	return &PowerGridRenderer{
		LineRenderer:  NewLineRenderer(mgl32.Vec3{0, 1, 0}),
		PlantRenderer: NewRegionRenderer(mgl32.Vec3{0.5, 0.5, 0.0})}
}
