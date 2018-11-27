package flat

import (
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/core/gamegrid"
	"go-simulate-a-city/sim/core/mailroom"
	"go-simulate-a-city/sim/engine/power"
	"go-simulate-a-city/sim/ui/lines"
	"go-simulate-a-city/sim/ui/region"

	"github.com/go-gl/mathgl/mgl32"
)

type PowerGridRenderer struct {
	powerRegions        []*commonMath.Region
	newPowerGridChannel chan *commonMath.Region

	offsetChangeChannel chan mgl32.Vec2
	cameraOffset        mgl32.Vec2
	scaleChangeChannel  chan float32
	cameraScale         float32
}

func NewPowerGridRenderer() *PowerGridRenderer {
	renderer := PowerGridRenderer{
		cameraOffset:        mgl32.Vec2{0, 0},
		cameraScale:         1.0,
		powerRegions:        make([]*commonMath.Region, 0),
		newPowerGridChannel: make(chan *commonMath.Region, 10),
		offsetChangeChannel: make(chan mgl32.Vec2, 10),
		scaleChangeChannel:  make(chan float32, 10)}

	mailroom.CameraOffsetRegChannel <- renderer.offsetChangeChannel
	mailroom.CameraScaleRegChannel <- renderer.scaleChangeChannel
	mailroom.NewPlantRegionRegChannel <- renderer.newPowerGridChannel
	return &renderer
}

func (p *PowerGridRenderer) drainInputChannels() {
	inputLeft := true
	for inputLeft {
		select {
		case p.cameraOffset = <-p.offsetChangeChannel:
		case p.cameraScale = <-p.scaleChangeChannel:
		case region := <-p.newPowerGridChannel:
			p.powerRegions = append(p.powerRegions, region)
		default:
			inputLeft = false
		}
	}
}

func (p *PowerGridRenderer) Render(shadingProgram *region.RegionShaderProgram) {
	p.drainInputChannels()

	shadingProgram.PreRender()
	for _, region := range p.powerRegions {
		mappedRegion := gamegrid.MapEngineRegionToScreen(region, p.cameraScale, p.cameraOffset)
		shadingProgram.Render(mappedRegion, mgl32.Vec3{0.5, 0.5, 0.0})
	}
}

func RenderPowerLines(grid *power.PowerGrid, camera *Camera, shadingProgram *lines.LinesShaderProgram) {
	shadingProgram.PreRender()

	lines := make([][2]mgl32.Vec2, 0)
	grid.IterateLines(func(line *power.PowerLine) {
		mappedLine := camera.MapEngineLineToScreen(line.GetLine())
		lines = append(lines, mappedLine)
		if len(lines) > 200 {
			shadingProgram.Render(lines, mgl32.Vec3{0, 1, 0})
			lines = make([][2]mgl32.Vec2, 0)
		}
	})

	if len(lines) > 0 {
		shadingProgram.Render(lines, mgl32.Vec3{0, 1, 0})
	}
}
