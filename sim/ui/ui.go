package ui

import (
	"go-simulate-a-city/sim/ui/lines"
	"go-simulate-a-city/sim/ui/overlay"
	"go-simulate-a-city/sim/ui/region"

	"github.com/go-gl/glfw/v3.2/glfw"
)

type UiInfrastructure struct {
	OverlayProgram *overlay.OverlayShaderProgram
	RegionProgram  *region.RegionShaderProgram
	LinesProgram   *lines.LinesShaderProgram
}

var Ui UiInfrastructure

// Defines common UI initialization, for both 2D and 3D rendering modes.
func Init(window *glfw.Window) {
	Ui.OverlayProgram = overlay.NewOverlayShaderProgram()
	Ui.RegionProgram = region.NewRegionShaderProgram()
	Ui.LinesProgram = lines.NewLinesShaderProgram()
}

func Delete() {
	Ui.OverlayProgram.Delete()
	Ui.RegionProgram.Delete()
	Ui.LinesProgram.Delete()
}
