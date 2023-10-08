package flat

import (
	"sim/core/dto/geometry"
	"sim/core/gamegrid"
	"sim/core/mailroom"
	"sim/ui"

	"github.com/go-gl/mathgl/mgl32"
)

// Defines how to render generic lines in a channel-based manner
type LineRenderer struct {
	offsetChangeChannel chan mgl32.Vec2
	scaleChangeChannel  chan float32

	cameraOffset mgl32.Vec2
	cameraScale  float32

	lineColor         mgl32.Vec3
	lastRenderedLines [][2]mgl32.Vec2
	lines             map[int64][2]mgl32.Vec2
	newInput          bool

	NewLineChannel    chan geometry.IdLine
	DeleteLineChannel chan int64
}

func NewLineRenderer(lineColor mgl32.Vec3) *LineRenderer {
	renderer := LineRenderer{
		offsetChangeChannel: make(chan mgl32.Vec2, 10),
		scaleChangeChannel:  make(chan float32, 10),
		cameraOffset:        mgl32.Vec2{0, 0},
		cameraScale:         1.0,
		lineColor:           lineColor,
		lastRenderedLines:   make([][2]mgl32.Vec2, 0),
		lines:               make(map[int64][2]mgl32.Vec2),
		newInput:            false,
		NewLineChannel:      make(chan geometry.IdLine, 50),
		DeleteLineChannel:   make(chan int64, 50)}

	mailroom.CameraOffsetRegChannel <- renderer.offsetChangeChannel
	mailroom.CameraScaleRegChannel <- renderer.scaleChangeChannel

	return &renderer
}

func (r *LineRenderer) drainInputChannels() {
	inputLeft := true
	r.newInput = false
	for inputLeft {
		select {
		case r.cameraOffset = <-r.offsetChangeChannel:
			r.newInput = true
		case r.cameraScale = <-r.scaleChangeChannel:
			r.newInput = true
		case deletionId := <-r.DeleteLineChannel:
			delete(r.lines, deletionId)
			r.newInput = true
		case idLine := <-r.NewLineChannel:
			r.lines[idLine.Id] = idLine.Line
			r.newInput = true
		default:
			inputLeft = false
		}
	}
}

func (r *LineRenderer) Render() {
	r.drainInputChannels()

	if r.newInput {
		r.lastRenderedLines = make([][2]mgl32.Vec2, 0)
		for _, line := range r.lines {
			mappedLine := [2]mgl32.Vec2{
				gamegrid.MapPositionToScreen(line[0], r.cameraScale, r.cameraOffset),
				gamegrid.MapPositionToScreen(line[1], r.cameraScale, r.cameraOffset)}
			r.lastRenderedLines = append(r.lastRenderedLines, mappedLine)
		}
	}

	// TODO: Update line renderer to support caching buffers,
	//  which will significantly improve no-op perf.
	if len(r.lastRenderedLines) > 0 {
		ui.Ui.LinesProgram.Render(r.lastRenderedLines, r.lineColor)
	}
}
