package input

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var InputBuffer *InputBufferAgent

// The input buffer agent separates the GLFW input system from a channel / agent
// based agent design.
type InputBufferAgent struct {
	ControlChannel chan int

	mouseMoveRegistrations   []chan mgl32.Vec2
	mouseScrollRegistrations []chan float32
	keyPressedRegistrations  []chan glfw.Key
	keyReleasedRegistrations []chan glfw.Key
	MouseMoveChannel         chan mgl32.Vec2
	MouseMoveRegChannel      chan chan mgl32.Vec2
	MouseScrollChannel       chan float32
	MouseScrollRegChannel    chan chan float32
	PressedKeysChannel       chan glfw.Key
	ReleasedKeysChannel      chan glfw.Key
	PressedKeysRegChannel    chan chan glfw.Key
	ReleasedKeysRegChannel   chan chan glfw.Key
}

func SetupInputBufferAgent() {
	agent := InputBufferAgent{
		ControlChannel:           make(chan int),
		mouseMoveRegistrations:   make([]chan mgl32.Vec2, 0),
		mouseScrollRegistrations: make([]chan float32, 0),
		keyPressedRegistrations:  make([]chan glfw.Key, 0),
		keyReleasedRegistrations: make([]chan glfw.Key, 0),
		MouseMoveChannel:         make(chan mgl32.Vec2),
		MouseMoveRegChannel:      make(chan chan mgl32.Vec2),
		MouseScrollChannel:       make(chan float32),
		MouseScrollRegChannel:    make(chan chan float32),
		PressedKeysChannel:       make(chan glfw.Key),
		ReleasedKeysChannel:      make(chan glfw.Key),
		PressedKeysRegChannel:    make(chan chan glfw.Key),
		ReleasedKeysRegChannel:   make(chan chan glfw.Key)}

	go agent.run()
	InputBuffer = &agent
}

func (i *InputBufferAgent) run() {
	for {
		select {
		case key := <-i.PressedKeysChannel:
			for _, agent := range i.keyPressedRegistrations {
				agent <- key
			}
			break
		case reg := <-i.PressedKeysRegChannel:
			i.keyPressedRegistrations = append(i.keyPressedRegistrations, reg)
			break
		case key := <-i.ReleasedKeysChannel:
			for _, agent := range i.keyReleasedRegistrations {
				agent <- key
			}
			break
		case reg := <-i.ReleasedKeysRegChannel:
			i.keyReleasedRegistrations = append(i.keyReleasedRegistrations, reg)
			break
		case input := <-i.MouseMoveChannel:
			for _, agent := range i.mouseMoveRegistrations {
				agent <- input
			}
			break
		case reg := <-i.MouseMoveRegChannel:
			i.mouseMoveRegistrations = append(i.mouseMoveRegistrations, reg)
			break
		case input := <-i.MouseScrollChannel:
			for _, agent := range i.mouseScrollRegistrations {
				agent <- input
			}
			break
		case reg := <-i.MouseScrollRegChannel:
			i.mouseScrollRegistrations = append(i.mouseScrollRegistrations, reg)
			break
		case _ = <-i.ControlChannel:
			return
		}
	}
}
