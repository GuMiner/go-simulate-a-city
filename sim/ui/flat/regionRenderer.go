package flat

import (
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/core/dto/geometry"
	"go-simulate-a-city/sim/core/gamegrid"
	"go-simulate-a-city/sim/core/mailroom"
	"go-simulate-a-city/sim/ui"

	"github.com/go-gl/mathgl/mgl32"
)

// Defines how to render generic regions in a channel-based manner
type RegionRenderer struct {
	offsetChangeChannel chan mgl32.Vec2
	scaleChangeChannel  chan float32

	cameraOffset mgl32.Vec2
	cameraScale  float32

	regionColor           mgl32.Vec3
	lastRendereredRegions []commonMath.Region
	regions               map[int64]commonMath.Region
	newInput              bool

	NewRegionChannel    chan geometry.IdRegion
	DeleteRegionChannel chan int64
}

func NewRegionRenderer(regionColor mgl32.Vec3) *RegionRenderer {
	renderer := RegionRenderer{
		offsetChangeChannel:   make(chan mgl32.Vec2, 10),
		scaleChangeChannel:    make(chan float32, 10),
		cameraOffset:          mgl32.Vec2{0, 0},
		cameraScale:           1.0,
		regionColor:           regionColor,
		lastRendereredRegions: make([]commonMath.Region, 0),
		regions:               make(map[int64]commonMath.Region),
		newInput:              false,
		NewRegionChannel:      make(chan geometry.IdRegion, 50),
		DeleteRegionChannel:   make(chan int64, 50)}

	mailroom.CameraOffsetRegChannel <- renderer.offsetChangeChannel
	mailroom.CameraScaleRegChannel <- renderer.scaleChangeChannel

	return &renderer
}

func (r *RegionRenderer) drainInputChannels() {
	inputLeft := true
	r.newInput = false
	for inputLeft {
		select {
		case r.cameraOffset = <-r.offsetChangeChannel:
			r.newInput = true
		case r.cameraScale = <-r.scaleChangeChannel:
			r.newInput = true
		case deletionId := <-r.DeleteRegionChannel:
			delete(r.regions, deletionId)
			r.newInput = true
		case idRegion := <-r.NewRegionChannel:
			if idRegion.Id == -1 {
				// Special case -- if someone sends an invalid ID, we reset EVERYTHING
				r.regions = make(map[int64]commonMath.Region)
			} else {
				r.regions[idRegion.Id] = idRegion.Region
			}

			r.newInput = true
		default:
			inputLeft = false
		}
	}
}

func (r *RegionRenderer) Render() {
	r.drainInputChannels()

	if r.newInput {
		r.lastRendereredRegions = make([]commonMath.Region, 0)
		for _, region := range r.regions {
			mappedRegion := gamegrid.MapEngineRegionToScreen(&region, r.cameraScale, r.cameraOffset)
			r.lastRendereredRegions = append(r.lastRendereredRegions, *mappedRegion)
		}
	}

	// TODO: Update region renderer to support caching buffers,
	//  which will significantly improve no-op perf.
	for _, region := range r.lastRendereredRegions {
		ui.Ui.RegionProgram.Render(&region, r.regionColor)
	}
}
