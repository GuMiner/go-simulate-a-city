package flat

import (
	"go-simulate-a-city/common/commonmath"
	"go-simulate-a-city/common/commonopengl"
	"time"

	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/input"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var isLeftPressed = false
var isRightPressed = false
var isUpPressed = false
var isDownPressed = false
var lastTimeTick time.Time
var ticked bool = false

type Camera struct {
	mouseMoves     chan mgl32.Vec2
	mouseScrolls   chan float32
	keyPresses     chan glfw.Key
	keyReleases    chan glfw.Key
	highResTicks   chan time.Time
	ControlChannel chan int

	offsetChangeRegs       []chan mgl32.Vec2
	OffsetChangeRegChannel chan chan mgl32.Vec2

	scaleChangeRegs       []chan float32
	ScaleChangeRegChannel chan chan float32

	zoomFactor float32
	offset     mgl32.Vec2

	mouseBoardPos mgl32.Vec2

	lastUpdateTicks uint
	keyMotionAmount float32
}

func NewCamera(
	mouseMoveRegChannel chan chan mgl32.Vec2,
	mouseScrollRegChannel chan chan float32,
	keyPressedRegChannel chan chan glfw.Key,
	keyReleasedRegChannel chan chan glfw.Key,
	highResRegChannel chan chan time.Time) *Camera {

	camera := Camera{
		zoomFactor:             1.0,
		offset:                 mgl32.Vec2{0, 0},
		mouseMoves:             make(chan mgl32.Vec2, 2),
		mouseScrolls:           make(chan float32, 2),
		keyPresses:             make(chan glfw.Key, 2),
		keyReleases:            make(chan glfw.Key, 2),
		highResTicks:           make(chan time.Time, 2),
		ControlChannel:         make(chan int),
		lastUpdateTicks:        0,
		offsetChangeRegs:       make([]chan mgl32.Vec2, 0),
		OffsetChangeRegChannel: make(chan chan mgl32.Vec2),
		scaleChangeRegs:        make([]chan float32, 0),
		ScaleChangeRegChannel:  make(chan chan float32)}

	mouseMoveRegChannel <- camera.mouseMoves
	mouseScrollRegChannel <- camera.mouseScrolls
	highResRegChannel <- camera.highResTicks
	keyPressedRegChannel <- camera.keyPresses
	keyReleasedRegChannel <- camera.keyReleases

	go camera.run()

	return &camera
}

func parseKeyCode(keyCode glfw.Key, stateTransition bool) {
	switch keyCode {
	case input.GetKeyCode(input.MoveUpKey):
		isUpPressed = stateTransition
		break
	case input.GetKeyCode(input.MoveRightKey):
		isRightPressed = stateTransition
		break
	case input.GetKeyCode(input.MoveDownKey):
		isDownPressed = stateTransition
		break
	case input.GetKeyCode(input.MoveLeftKey):
		isLeftPressed = stateTransition
		break
	default:
		break
	}
}

func (c *Camera) handleTickMotion(interval float32) {
	keyMotionAmount := interval * config.Config.Ui.Camera.KeyMotionFactor * (1.0 / c.zoomFactor)
	offsetChanged := false
	if isLeftPressed {
		c.offset[0] -= keyMotionAmount
		offsetChanged = true
	}

	if isRightPressed {
		c.offset[0] += keyMotionAmount
		offsetChanged = true
	}

	if isUpPressed {
		c.offset[1] -= keyMotionAmount
		offsetChanged = true
	}

	if isDownPressed {
		c.offset[1] += keyMotionAmount
		offsetChanged = true
	}

	if offsetChanged {
		for _, reg := range c.offsetChangeRegs {
			reg <- c.offset
		}
	}
}

func (c *Camera) run() {
	for {
		select {
		case reg := <-c.OffsetChangeRegChannel:
			c.offsetChangeRegs = append(c.offsetChangeRegs, reg)
			break
		case reg := <-c.ScaleChangeRegChannel:
			c.scaleChangeRegs = append(c.scaleChangeRegs, reg)
			break
		case mousePos := <-c.mouseMoves:
			c.mouseBoardPos = c.MapPixelPosToBoard(mousePos)
			break
		case scrollAmount := <-c.mouseScrolls:
			c.zoomFactor *= (1.0 + scrollAmount*config.Config.Ui.Camera.MouseScrollFactor)
			break
		case keyCode := <-c.keyPresses:
			parseKeyCode(keyCode, true)
			break
		case keyCode := <-c.keyReleases:
			parseKeyCode(keyCode, false)
		case time := <-c.highResTicks:
			if !ticked {
				ticked = true
			} else {
				timeSinceLastTick := float32(time.Sub(lastTimeTick).Seconds())
				c.handleTickMotion(timeSinceLastTick)
			}

			lastTimeTick = time
			break
		case _ = <-c.ControlChannel:
			return
		}
	}
}

func (c *Camera) getMinMaxVisibleRange() (minTile mgl32.Vec2, maxTile mgl32.Vec2) {
	regionSize := config.Config.Terrain.RegionSize

	minTile = c.MapToBoard(mgl32.Vec2{0, 0}).Mul(1.0 / float32(regionSize))
	maxTile = c.MapToBoard(mgl32.Vec2{1, 1}).Mul(1.0 / float32(regionSize))
	return minTile, maxTile
}

func (c *Camera) ComputeVisibleRegions() []commonMath.IntVec2 {
	minTile, maxTile := c.getMinMaxVisibleRange()

	visibleTiles := make([]commonMath.IntVec2, 0)
	for i := int(minTile.X() - 1.0); i <= int(maxTile.X()+1.0); i++ {
		for j := int(minTile.Y() - 1.0); j <= int(maxTile.Y()+1.0); j++ {
			visibleTiles = append(visibleTiles, commonMath.IntVec2{i, j})
		}
	}

	return visibleTiles
}

func (c *Camera) ComputePrecacheRegions() []commonMath.IntVec2 {
	minTile, maxTile := c.getMinMaxVisibleRange()

	visibleTiles := make([]commonMath.IntVec2, 0)
	for i := int(minTile.X() - 2.0); i <= int(maxTile.X()+2.0); i++ {
		for j := int(minTile.Y() - 2.0); j <= int(maxTile.Y()+2.0); j++ {
			if i == int(minTile.X()-2.0) ||
				i == int(minTile.X()+2.0) ||
				j == int(minTile.Y()-2.0) ||
				j == int(maxTile.Y()+2.0) {
				visibleTiles = append(visibleTiles, commonMath.IntVec2{i, j})
			}
		}
	}

	return visibleTiles
}

// Maps a position in pixels to the board
func (c *Camera) MapPixelPosToBoard(pixelPos mgl32.Vec2) mgl32.Vec2 {
	windowSize := commonOpenGl.GetWindowSize()
	return c.MapToBoard(mgl32.Vec2{pixelPos.X() / windowSize.X(), pixelPos.Y() / windowSize.Y()})
}

// Maps a (0, 0) to (1, 1) screen position to a board location.
func (c *Camera) MapToBoard(screenPos mgl32.Vec2) mgl32.Vec2 {
	windowSize := commonOpenGl.GetWindowSize()

	modifiedRegionPos := mgl32.Vec2{(screenPos.X() - 0.5) * windowSize.X(), (screenPos.Y() - 0.5) * windowSize.Y()}
	regionPos := modifiedRegionPos.Mul(1.0 / c.zoomFactor).Add(c.offset)

	return regionPos
}

// Maps a region on the board to a GLSL (-1, -1) to (1, 1) region
func (c *Camera) MapEngineRegionToScreen(region *commonMath.Region) *commonMath.Region {
	// The only variables that are updated (for now) are position and scale
	region.Scale *= (1.0 / c.zoomFactor)
	region.Position = c.MapPositionToScreen(region.Position)
	return region
}

func (c *Camera) MapEngineLineToScreen(line [2]mgl32.Vec2) [2]mgl32.Vec2 {
	return [2]mgl32.Vec2{
		c.MapPositionToScreen(line[0]),
		c.MapPositionToScreen(line[1])}
}

func (c *Camera) MapPositionToScreen(point mgl32.Vec2) mgl32.Vec2 {
	windowSize := commonOpenGl.GetWindowSize()
	point = point.Sub(c.offset).Mul(c.zoomFactor)
	point = mgl32.Vec2{2 * point.X() / windowSize.X(), -2 * point.Y() / windowSize.Y()}
	return point
}

// Resizes a full-size region to the appropriate scale given the current screen size and zoom factor
// Returns the screen size (a full size tile will span from (0, 0) to (1, 1))
func GetRegionScale(zoomFactor float32) mgl32.Vec2 {
	regionSize := config.Config.Terrain.RegionSize
	windowSize := commonOpenGl.GetWindowSize()
	return mgl32.Vec2{
		zoomFactor * float32(regionSize) / windowSize.X(),
		zoomFactor * float32(regionSize) / windowSize.Y()}
}

// Returns the screen position ((0, 0) to (1, 1)) of the region tile requested
func GetRegionOffset(x, y int, offset mgl32.Vec2, zoomFactor float32) mgl32.Vec2 {
	regionSize := config.Config.Terrain.RegionSize
	windowSize := commonOpenGl.GetWindowSize()

	regionStart := mgl32.Vec2{float32(x * regionSize), float32(y * regionSize)}
	modifiedRegionStart := regionStart.Sub(offset).Mul(zoomFactor)

	return mgl32.Vec2{modifiedRegionStart.X()/windowSize.X() + 0.5, modifiedRegionStart.Y()/windowSize.Y() + 0.5}
}

func (c *Camera) GetZoomFactor() float32 {
	return c.zoomFactor
}

func (c *Camera) GetOffset() mgl32.Vec2 {
	return c.offset
}

func (c *Camera) getScaleMotionFactor() float32 {
	return c.zoomFactor
}
