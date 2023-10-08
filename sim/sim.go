package main

import (
	commonColor "common/commoncolor"
	commonConfig "common/commonconfig"
	commonOpenGl "common/commonopengl"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sim/config"
	"sim/core/mailroom"
	"sim/engine"
	"sim/engine/core"
	"sim/input"
	"sim/input/editorEngine"
	"sim/ui"
	"sim/ui/flat"
	"time"

	"github.com/go-gl/gl/v4.4-core/gl"
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
	// Navigate to http://localhost:8765/debug/pprof/goroutine?debug=1 to see the current goroutines
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

	mailroom.Init()

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
	mailroom.EngineCancelChannel = editorEngine.CancellationRegChannel

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

	vehicleRenderer := flat.NewVehicleRenderer()
	mailroom.NewRoadLineIdChannel = vehicleRenderer.RoadLineRegChannel
	mailroom.NewRoadTerminusChannel = vehicleRenderer.TerminusChannel
	mailroom.VehicleUpdateChannel = vehicleRenderer.VehicleUpdateChannel

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
		vehicleRenderer.Renderer.Render()

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
