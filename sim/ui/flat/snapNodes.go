package flat

import (
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/core/dto/geometry"

	"github.com/go-gl/mathgl/mgl32"
)

type SnapRenderer struct {
	SnappedNodesUpdateChannel chan []mgl32.Vec2
	NodeRenderer              *RegionRenderer
}

func NewSnapRenderer() *SnapRenderer {
	renderer := SnapRenderer{
		SnappedNodesUpdateChannel: make(chan []mgl32.Vec2),
		NodeRenderer:              NewRegionRenderer(mgl32.Vec3{0.0, 1.0, 0.0})}

	go renderer.run()
	return &renderer
}

func (r *SnapRenderer) run() {
	for {
		positions := <-r.SnappedNodesUpdateChannel

		// Reset the renderer
		r.NodeRenderer.NewRegionChannel <- geometry.NewIdRegion(-1, commonMath.Region{})
		for idx, pos := range positions {
			region := commonMath.Region{
				RegionType:  commonMath.CircleRegion,
				Position:    pos,
				Scale:       50,
				Orientation: 0}

			r.NodeRenderer.NewRegionChannel <- geometry.NewIdRegion(int64(idx), region)
		}
	}
}
