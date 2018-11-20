package flat

import (
	"go-simulate-a-city/sim/core/gamegrid"
	"go-simulate-a-city/sim/engine/terrain"
	"go-simulate-a-city/sim/ui"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type TerrainOverlayManager struct {
	offsetChangeChannel chan mgl32.Vec2
	scaleChangeChannel  chan float32
	newTerrainChannel   chan *terrain.TerrainUpdate

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

func NewTerrainOverlayManager(
	offsetChangeRegChannel chan chan mgl32.Vec2,
	scaleChangeRegChannel chan chan float32,
	newTerrainRegChannel chan chan *terrain.TerrainUpdate) *TerrainOverlayManager {
	manager := TerrainOverlayManager{
		cameraOffset:        mgl32.Vec2{0, 0},
		cameraScale:         1.0,
		offsetChangeChannel: make(chan mgl32.Vec2, 10),
		scaleChangeChannel:  make(chan float32, 10),
		newTerrainChannel:   make(chan *terrain.TerrainUpdate, 10),
		TerrainOverlays:     make(map[int]map[int]*TerrainOverlay)}

	offsetChangeRegChannel <- manager.offsetChangeChannel
	scaleChangeRegChannel <- manager.scaleChangeChannel
	newTerrainRegChannel <- manager.newTerrainChannel

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
		case newTerrain := <-t.newTerrainChannel:
			t.GetOrAddTerrainOverlay(
				newTerrain.Pos.X(),
				newTerrain.Pos.Y()).SetTerrain(newTerrain.Texels)
			break
		default:
			inputLeft = false
			break
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
