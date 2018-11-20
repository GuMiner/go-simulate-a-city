package main

import (
	"go-simulate-a-city/common/commoncolor"
	"go-simulate-a-city/common/commonconfig"
	"go-simulate-a-city/common/commonopengl"
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/engine"
	"go-simulate-a-city/sim/engine/core"
	"go-simulate-a-city/sim/engine/terrain"
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
	"github.com/go-gl/mathgl/mgl32"
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
	setInputCallbacks(window)
	commonOpenGl.ConfigureOpenGl()

	commonColor.InitializeColorGradient(
		commonConfig.Config.ColorGradient.Steps,
		commonConfig.Config.ColorGradient.Saturation,
		commonConfig.Config.ColorGradient.Luminosity)

	ui.Init(window)
	defer ui.Delete()

	editorEngine.Init()

	core.Init()
	camera := flat.NewCamera(
		input.InputBuffer.MouseMoveRegChannel,
		input.InputBuffer.MouseScrollRegChannel,
		input.InputBuffer.PressedKeysRegChannel,
		input.InputBuffer.ReleasedKeysRegChannel,
		core.CoreTimer.HighResRegChannel)

	// Setup simulation

	engine := engine.NewEngine(camera.OffsetChangeRegChannel, camera.ScaleChangeRegChannel)

	terrainOverlayManager := flat.NewTerrainOverlayManager(
		camera.OffsetChangeRegChannel,
		camera.ScaleChangeRegChannel,
		engine.GetTerrainMap().NewTerrainRegChannel)
	defer terrainOverlayManager.Delete()

	// Precaching only does the outer borders. Ensure we start with a clean state.
	engine.GetTerrainMap().ControlChannel <- terrain.FORCE_REFRESH

	startTime := time.Now()
	frameTime := float32(0.1)
	lastElapsed := float32(0.0)
	elapsed := lastElapsed

	paused := false
	update := func() {
		lastElapsed = elapsed
		elapsed = float32(time.Since(startTime)) / float32(time.Second)
		frameTime = elapsed - lastElapsed

		// Must be first.
		glfw.PollEvents()

		editorStateUpdated, editorSubStateUpdated := editorEngine.Update()
		if editorStateUpdated || editorSubStateUpdated {
			// The edit state has updated, update as needed
			ui.UpdateEditorState(editorEngine.EngineState, window)
		}

		// TODO -- this all should be event / message based.
		// To avoid blocking with what exists now, getting the region map runs on its own goroutine
		// Load new terrain regions based on what is visible.
		// engine.PrecacheRegions(camera.ComputePrecacheRegions())

		boardPos := camera.MapPixelPosToBoard(mgl32.Vec2{0, 0}) //input.MousePos)
		if editorStateUpdated || true {                         // mouseMoved {
			engine.Hypotheticals.ComputeHypotheticalRegion(engine, &editorEngine.EngineState)
			engine.ComputeSnapNodes(&editorEngine.EngineState)
		}

		if input.MousePressEvent {
			engine.MousePress(boardPos, editorEngine.EngineState)
			input.MousePressEvent = false
		}

		if input.MouseReleaseEvent {
			engine.MouseRelease(boardPos, editorEngine.EngineState)
			input.MouseReleaseEvent = false
		}

		if true { // mouseMoved {
			engine.MouseMoved(boardPos, editorEngine.EngineState)
		}

		if input.IsTyped(input.CancelKey) {
			engine.CancelState(editorEngine.EngineState)
		}

		if input.IsTyped(input.PauseKey) {
			paused = !paused
		}

		if !paused {
			engine.StepSim(frameTime)
		}
		engine.StepEdit(frameTime, editorEngine.EngineState)
	}

	render := func() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// Render each visible region
		ui.Ui.OverlayProgram.PreRender()
		terrainOverlayManager.Render()

		ui.Ui.RegionProgram.PreRender()
		for _, hypotheticalRegion := range engine.Hypotheticals.Regions {
			mappedRegion := camera.MapEngineRegionToScreen(&hypotheticalRegion.Region)
			ui.Ui.RegionProgram.Render(mappedRegion, hypotheticalRegion.Color)
		}

		ui.Ui.LinesProgram.PreRender()
		for _, hypotheticalLine := range engine.Hypotheticals.Lines {
			mappedLine := camera.MapEngineLineToScreen(hypotheticalLine.Line)
			ui.Ui.LinesProgram.Render([][2]mgl32.Vec2{mappedLine}, hypotheticalLine.Color)
		}

		flat.RenderPowerPlants(engine.GetPowerGrid(), camera, ui.Ui.RegionProgram)
		flat.RenderPowerLines(engine.GetPowerGrid(), camera, ui.Ui.LinesProgram)

		flat.RenderRoadLines(engine.GetRoadGrid(), camera, ui.Ui.LinesProgram)

		flat.RenderSnapNodes(engine.GetSnapElements(), camera, ui.Ui.RegionProgram)
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
