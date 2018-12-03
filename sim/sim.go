package main

import (
	"go-simulate-a-city/common/commoncolor"
	"go-simulate-a-city/common/commonconfig"
	"go-simulate-a-city/common/commonopengl"
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/core/mailroom"
	"go-simulate-a-city/sim/engine"
	"go-simulate-a-city/sim/engine/core"
	"go-simulate-a-city/sim/input"
	"go-simulate-a-city/sim/input/editorEngine"
	"go-simulate-a-city/sim/ui"
	"go-simulate-a-city/sim/ui/flat"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

func init() {
	runtime.LockOSThread()
}

func setInputCallbacks(window *glfw.Window) {
	window.SetFramebufferSizeCallback(commonOpenGl.ResizeViewport)

	input.CreateDefaultKeyMap()
	window.SetCursorPosCallback(input.HandleMouseMove)
	window.SetMouseButtonCallback(input.HandleMouseButton)
	window.SetScrollCallback(input.HandleMouseScroll)
	window.SetKeyCallback(input.HandleKeyInput)
}

func main() {
	// Start memory profiling
	go func() {
		log.Println("Starting performance diagnostics on localhost:8765...")
		log.Println(http.ListenAndServe("localhost:8765", nil))
	}()

	config.Load("./data/config/", "./data/commonConfig.json")

	commonOpenGl.InitGlfw()
	defer glfw.Terminate()

	commonOpenGl.InitViewport()
	window, err := glfw.CreateWindow(
		int(commonOpenGl.GetWindowSize().X()),
		int(commonOpenGl.GetWindowSize().Y()),
		commonConfig.Config.Window.Title, nil, nil)
	if err != nil {
		panic(err)
	}

	window.MakeContextCurrent()

	input.SetupInputBufferAgent()
	mailroom.MousePressedRegChannel = input.InputBuffer.MousePressedRegChannel
	mailroom.MouseReleasedRegChannel = input.InputBuffer.MouseReleasedRegChannel

	setInputCallbacks(window)
	commonOpenGl.ConfigureOpenGl()

	commonColor.InitializeColorGradient(
		commonConfig.Config.ColorGradient.Steps,
		commonConfig.Config.ColorGradient.Saturation,
		commonConfig.Config.ColorGradient.Luminosity)

	editorEngine := editorEngine.NewEditorEngine(input.InputBuffer.PressedKeysRegChannel)
	mailroom.EngineModeRegChannel = editorEngine.EngineModeRegChannel
	mailroom.EngineAddModeRegChannel = editorEngine.EngineAddModeRegChannel
	mailroom.EngineDrawModeRegChannel = editorEngine.EngineDrawModeRegChannel
	mailroom.SnapSettingsRegChannel = editorEngine.SnapSettingsRegChannel

	ui.Init(window)
	customCursors := ui.NewCustomCursors()
	defer customCursors.Delete()
	defer ui.Delete()

	core.Init()
	camera := flat.NewCamera(
		input.InputBuffer.MouseMoveRegChannel,
		input.InputBuffer.MouseScrollRegChannel,
		input.InputBuffer.PressedKeysRegChannel,
		input.InputBuffer.ReleasedKeysRegChannel)

	mailroom.CameraOffsetRegChannel = camera.OffsetChangeRegChannel
	mailroom.CameraScaleRegChannel = camera.ScaleChangeRegChannel
	mailroom.BoardPosChangeRegChannel = camera.BoardPosRegChannel

	// Setup simulation
	_ = engine.NewEngine()

	powerGridRenderer := flat.NewPowerGridRenderer()
	mailroom.NewPowerLineChannel = powerGridRenderer.LineRenderer.NewLineChannel
	mailroom.DeletePowerLineChannel = powerGridRenderer.LineRenderer.DeleteLineChannel

	mailroom.NewPowerPlantChannel = powerGridRenderer.PlantRenderer.NewRegionChannel
	mailroom.DeletePowerPlantChannel = powerGridRenderer.PlantRenderer.DeleteRegionChannel

	roadGridRenderer := flat.NewRoadGridRenderer()
	mailroom.NewRoadLineChannel = roadGridRenderer.Renderer.NewLineChannel
	mailroom.DeleteRoadLineChannel = roadGridRenderer.Renderer.DeleteLineChannel

	snapRenderer := flat.NewSnapRenderer()
	mailroom.SnappedNodesUpdateChannel = snapRenderer.SnappedNodesUpdateChannel

	terrainOverlayManager := flat.NewTerrainOverlayManager()
	defer terrainOverlayManager.Delete()

	// paused := false

	startTime := time.Now()
	frameTime := float32(0.1)
	lastElapsed := float32(0.0)
	elapsed := lastElapsed
	update := func() {
		lastElapsed = elapsed
		elapsed = float32(time.Since(startTime)) / float32(time.Second)
		frameTime = elapsed - lastElapsed

		// Must be first.
		glfw.PollEvents()

		camera.StepUpdate(frameTime)
		customCursors.Update(window)

		//
		// if input.IsTyped(input.CancelKey) {
		// 	engine.CancelState(editorEngine.EngineState)
		// }
		//
		// if input.IsTyped(input.PauseKey) {
		// 	paused = !paused
		// }

		// engine.StepEdit(frameTime, editorEngine.EngineState)
	}

	render := func() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render each visible region
		ui.Ui.OverlayProgram.PreRender()
		terrainOverlayManager.Render()

		ui.Ui.RegionProgram.PreRender()

		powerGridRenderer.PlantRenderer.Render()
		snapRenderer.NodeRenderer.Render()
		// for _, hypotheticalRegion := range engine.Hypotheticals.Regions {
		// 	mappedRegion := camera.MapEngineRegionToScreen(&hypotheticalRegion.Region)
		// 	ui.Ui.RegionProgram.Render(mappedRegion, hypotheticalRegion.Color)
		// }

		ui.Ui.LinesProgram.PreRender()
		// for _, hypotheticalLine := range engine.Hypotheticals.Lines {
		// 	mappedLine := camera.MapEngineLineToScreen(hypotheticalLine.Line)
		// 	ui.Ui.LinesProgram.Render([][2]mgl32.Vec2{mappedLine}, hypotheticalLine.Color)
		// }

		roadGridRenderer.Renderer.Render()
		powerGridRenderer.LineRenderer.Render()
	}

	RenderLoop(update, render, window)
}

func RenderLoop(update, render func(), window *glfw.Window) {
	for !window.ShouldClose() {
		update()

		// Render the full display.
		commonOpenGl.ResetViewport()
		render()
		window.SwapBuffers()
	}
}
