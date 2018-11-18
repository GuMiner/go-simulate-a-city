package flat

import (
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/sim/ui"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type TerrainOverlayManager struct {
	offsetChangeChannel chan mgl32.Vec2
	scaleChangeChannel  chan float32

	cameraOffset    mgl32.Vec2
	cameraScale     float32
	activeOverlays  []commonMath.IntVec2
	TerrainOverlays map[int]map[int]*TerrainOverlay
}

func (t *TerrainOverlayManager) GetOrAddTerrainOverlay(x, y int) *TerrainOverlay {
	if _, ok := t.TerrainOverlays[x]; !ok {
		t.TerrainOverlays[x] = make(map[int]*TerrainOverlay)
	}

	if _, ok := t.TerrainOverlays[x][y]; !ok {
		var textureId uint32
		gl.GenTextures(1, &textureId)
		t.TerrainOverlays[x][y] = NewTerrainOverlay(textureId)
	}

	return t.TerrainOverlays[x][y]
}

func NewTerrainOverlayManager(
	offsetChangeRegChannel chan chan mgl32.Vec2,
	scaleChangeRegChannel chan chan float32) *TerrainOverlayManager {
	manager := TerrainOverlayManager{
		offsetChangeChannel: make(chan mgl32.Vec2, 10),
		scaleChangeChannel:  make(chan float32, 10),
		activeOverlays:      make([]commonMath.IntVec2, 0),
		TerrainOverlays:     make(map[int]map[int]*TerrainOverlay)}

	offsetChangeRegChannel <- manager.offsetChangeChannel
	scaleChangeRegChannel <- manager.scaleChangeChannel

	return &manager
}

func (t *TerrainOverlayManager) drainInputChannels() {
	inputLeft := true
	for inputLeft {
		select {
		case t.cameraOffset = <-t.offsetChangeChannel:
			break
		case t.cameraScale = <-t.scaleChangeChannel:
			break
		default:
			inputLeft = false
			break
		}
	}
}

func (t *TerrainOverlayManager) Render() {
	t.drainInputChannels()

	for _, region := range t.activeOverlays {
		overlay := t.TerrainOverlays[region.X()][region.Y()]
		overlay.UpdateCameraOffset(region.X(), region.Y(), t.cameraOffset, t.cameraScale)
		ui.Ui.OverlayProgram.Render(overlay.GetOverlay())
	}
}

func (t *TerrainOverlayManager) Delete() {
	for _, value := range t.TerrainOverlays {
		for _, overlay := range value {
			gl.DeleteTextures(1, &overlay.textureId)
		}
	}
}
