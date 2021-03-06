package editorengdto

type EditorMode int

const (
	Select EditorMode = iota
	Add
	Draw
)

type EditorAddMode int

const (
	PowerPlant EditorAddMode = iota
	PowerLine
	RoadLine
)

type ItemSubSelection int

const (
	Item1 ItemSubSelection = 0
	Item2 ItemSubSelection = 1
	Item3 ItemSubSelection = 2
	Item4 ItemSubSelection = 3
	Item5 ItemSubSelection = 4
	Item6 ItemSubSelection = 5
)

type EditorDrawMode int

const (
	TerrainFlatten EditorDrawMode = iota
	TerrainSharpen
	TerrainTrees
	TerrainShrubs
	TerrainHills
	TerrainValleys
)

type SnapToggle int

const (
	SnapToAngle SnapToggle = iota
	SnapToGrid
	SnapToElements
)

type SnapSetting struct {
	Setting SnapToggle
	State   bool
}
