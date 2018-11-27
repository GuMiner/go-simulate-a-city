package flat

import (
	"go-simulate-a-city/common/commonopengl"
	"time"

	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/core/gamegrid"
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

	boardPosRegs       []chan mgl32.Vec2
	BoardPosRegChannel chan chan mgl32.Vec2

	zoomFactor float32
	offset     mgl32.Vec2

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
		ScaleChangeRegChannel:  make(chan chan float32),
		boardPosRegs:           make([]chan mgl32.Vec2, 0),
		BoardPosRegChannel:     make(chan chan mgl32.Vec2)}

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
		case reg := <-c.BoardPosRegChannel:
			c.boardPosRegs = append(c.boardPosRegs, reg)
		case reg := <-c.OffsetChangeRegChannel:
			c.offsetChangeRegs = append(c.offsetChangeRegs, reg)
		case reg := <-c.ScaleChangeRegChannel:
			c.scaleChangeRegs = append(c.scaleChangeRegs, reg)
		case mousePos := <-c.mouseMoves:
			boardPos := c.MapPixelPosToBoard(mousePos)
			for _, reg := range c.boardPosRegs {
				reg <- boardPos
			}
		case scrollAmount := <-c.mouseScrolls:
			c.zoomFactor *= (1.0 + scrollAmount*config.Config.Ui.Camera.MouseScrollFactor)
			for _, reg := range c.scaleChangeRegs {
				reg <- c.zoomFactor
			}
		case keyCode := <-c.keyPresses:
			parseKeyCode(keyCode, true)
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
		case _ = <-c.ControlChannel:
			return
		}
	}
}

// Maps a position in pixels to the board
func (c *Camera) MapPixelPosToBoard(pixelPos mgl32.Vec2) mgl32.Vec2 {
	windowSize := commonOpenGl.GetWindowSize()
	return gamegrid.MapToBoard(
		mgl32.Vec2{pixelPos.X() / windowSize.X(), pixelPos.Y() / windowSize.Y()},
		c.offset,
		c.zoomFactor)
}

func (c *Camera) MapEngineLineToScreen(line [2]mgl32.Vec2) [2]mgl32.Vec2 {
	return [2]mgl32.Vec2{
		gamegrid.MapPositionToScreen(line[0], c.zoomFactor, c.offset),
		gamegrid.MapPositionToScreen(line[1], c.zoomFactor, c.offset)}
}
