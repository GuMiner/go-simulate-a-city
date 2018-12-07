package power

import (
	"fmt"
	"go-simulate-a-city/sim/core/dto/geometry"
	"go-simulate-a-city/sim/core/graph"
	"go-simulate-a-city/sim/core/mailroom"
	"go-simulate-a-city/sim/engine/finder"

	"github.com/go-gl/mathgl/mgl32"
)

type PowerGrid struct {
	finder *finder.ElementFinder
	grid   *graph.Graph
}

func NewPowerGrid(finder *finder.ElementFinder) *PowerGrid {
	grid := PowerGrid{
		finder: finder,
		grid:   graph.NewGraph()}
	return &grid
}

func (p *PowerGrid) Add(pos mgl32.Vec2, plantType string, plantSize PowerPlantSize) *PowerPlant {
	output, size := GetPowerOutputAndSize(plantType, plantSize)

	plant := PowerPlant{
		location:    pos,
		plantType:   plantType,
		namedSize:   plantSize,
		size:        float32(size),
		orientation: 0, // TODO: Rotation
		output:      output}

	gridId := p.grid.AddNode(&plant)
	fmt.Printf("Added power plant '%v'.\n", plant)

	p.finder.AddElementChannel <- finder.NewElement(gridId, finder.PowerTerminus, []mgl32.Vec2{pos})
	mailroom.NewPowerPlantChannel <- geometry.NewIdRegion(gridId, *plant.GetRegion())

	return &plant
}

// Adds a powerline. For both startNode and endNode, if -1 generates a new grid node, else uses an existing node.
// Returns the start ID, line ID, and end ID, in that order.
func (p *PowerGrid) AddLine(start, end mgl32.Vec2, capacity int64, startNode, endNode int64) (int64, int64, int64) {
	line := PowerLine{capacity: capacity}

	if startNode == endNode && startNode != -1 {
		fmt.Printf("Powerlines must be between nodes and cannot (for a single line) loop\n")
		return -1, -1, -1
	} else if startNode != -1 && endNode != -1 {
		// This might be a duplicate line.
		connectionStatus := p.grid.AddConnection(startNode, endNode, &line)
		if connectionStatus.Status == graph.Exists {
			fmt.Printf("There already is a line from %v to %v.\n", startNode, endNode)
			return -1, -1, -1
		} else {
			mailroom.NewPowerLineChannel <- geometry.NewIdLine(connectionStatus.Id, [2]mgl32.Vec2{start, end})
		}
	}

	if startNode == -1 {
		startNode = p.grid.AddNode(&PowerTerminus{location: start})
		p.finder.AddElementChannel <- finder.NewElement(startNode, finder.PowerTerminus, []mgl32.Vec2{start})
	}

	if endNode == -1 {
		endNode = p.grid.AddNode(&PowerTerminus{location: end})
		p.finder.AddElementChannel <- finder.NewElement(endNode, finder.PowerTerminus, []mgl32.Vec2{end})
	}

	connectionStatus := p.grid.AddConnection(startNode, endNode, &line)
	mailroom.NewPowerLineChannel <- geometry.NewIdLine(connectionStatus.Id, [2]mgl32.Vec2{start, end})

	return startNode, connectionStatus.Id, endNode
}
