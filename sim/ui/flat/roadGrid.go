package flat

import (
	"github.com/go-gl/mathgl/mgl32"
)

type RoadGridRenderer struct {
	Renderer *LineRenderer
}

func NewRoadGridRenderer() *RoadGridRenderer {
	return &RoadGridRenderer{
		Renderer: NewLineRenderer(mgl32.Vec3{1, 0, 0})}
}
