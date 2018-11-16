package input

import (
	"github.com/go-gl/mathgl/mgl32"
)

var InputBuffer *InputBufferAgent

// The input buffer agent separates the GLFW input system from a channel / agent
// based agent design.
type InputBufferAgent struct {
	ControlChannel chan int

	mouseMoveRegistrations []chan mgl32.Vec2
	MouseMoveChannel       chan mgl32.Vec2
	MouseMoveRegChannel    chan chan mgl32.Vec2
}

func SetupInputBufferAgent() {
	agent := InputBufferAgent{
		ControlChannel:         make(chan int),
		mouseMoveRegistrations: make([]chan mgl32.Vec2, 0),
		MouseMoveChannel:       make(chan mgl32.Vec2),
		MouseMoveRegChannel:    make(chan chan mgl32.Vec2)}

	go agent.run()
	InputBuffer = &agent
}

func (i *InputBufferAgent) run() {
	for {
		select {
		case input := <-i.MouseMoveChannel:
			for _, agent := range i.mouseMoveRegistrations {
				agent <- input
			}
		case reg := <-i.MouseMoveRegChannel:
			i.mouseMoveRegistrations = append(i.mouseMoveRegistrations, reg)
			break
		case _ = <-i.ControlChannel:
			return
		}
	}
}
