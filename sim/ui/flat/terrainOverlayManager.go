package flat

import (
	"sim/core/dto/terraindto"
	"sim/core/gamegrid"
	"sim/core/mailroom"
	"sim/ui"

	"github.com/go-gl/gl/v4.4-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type TerrainOverlayManager struct {
	offsetChangeChannel chan mgl32.Vec2
	scaleChangeChannel  chan float32
	newTerrainChannel   chan *terraindto.TerrainUpdate

	cameraOffset    mgl32.Vec2
	cameraScale     float32
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

func NewTerrainOverlayManager() *TerrainOverlayManager {
	manager := TerrainOverlayManager{
		cameraOffset:        mgl32.Vec2{0, 0},
		cameraScale:         1.0,
		offsetChangeChannel: make(chan mgl32.Vec2, 10),
		scaleChangeChannel:  make(chan float32, 10),
		newTerrainChannel:   make(chan *terraindto.TerrainUpdate, 10),
		TerrainOverlays:     make(map[int]map[int]*TerrainOverlay)}

	mailroom.CameraOffsetRegChannel <- manager.offsetChangeChannel
	mailroom.CameraScaleRegChannel <- manager.scaleChangeChannel
	mailroom.NewTerrainRegChannel <- manager.newTerrainChannel

	return &manager
}

func (t *TerrainOverlayManager) drainInputChannels() {
	inputLeft := true
	for inputLeft {
		select {
		case t.cameraOffset = <-t.offsetChangeChannel:
		case t.cameraScale = <-t.scaleChangeChannel:
		case newTerrain := <-t.newTerrainChannel:
			t.GetOrAddTerrainOverlay(
				newTerrain.Pos.X(),
				newTerrain.Pos.Y()).SetTerrain(newTerrain.Texels)
		default:
			inputLeft = false
		}
	}
}

func (t *TerrainOverlayManager) Render() {
	t.drainInputChannels()
	visibleRegions := gamegrid.ComputeVisibleRegions(t.cameraOffset, t.cameraScale)
	for _, region := range visibleRegions {
		overlay := t.GetOrAddTerrainOverlay(region.X(), region.Y())
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
