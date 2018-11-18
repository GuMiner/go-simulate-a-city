package input

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
)

var keyEvent bool = false
var typedKeys map[glfw.Key]bool = make(map[glfw.Key]bool)

var MousePressEvent bool = false
var MouseReleaseEvent bool = false
var PressedButtons map[glfw.MouseButton]bool = make(map[glfw.MouseButton]bool)

func HandleMouseMove(window *glfw.Window, xPos float64, yPos float64) {
	InputBuffer.MouseMoveChannel <- mgl32.Vec2{float32(xPos), float32(yPos)}
}

func HandleMouseScroll(window *glfw.Window, xOffset, yOffset float64) {
	InputBuffer.MouseScrollChannel <- float32(yOffset)
}

func HandleMouseButton(window *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		PressedButtons[button] = true
		MousePressEvent = true
	case glfw.Release:
		PressedButtons[button] = false
		MouseReleaseEvent = true
	}
}

func HandleKeyInput(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		InputBuffer.PressedKeysChannel <- key
		typedKeys[key] = true
	case glfw.Release:
		InputBuffer.ReleasedKeysChannel <- key
	}
}
