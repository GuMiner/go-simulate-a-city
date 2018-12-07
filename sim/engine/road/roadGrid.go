package road

import (
	"fmt"
	"go-simulate-a-city/sim/core/dto/geometry"
	"go-simulate-a-city/sim/core/graph"
	"go-simulate-a-city/sim/core/mailroom"
	"go-simulate-a-city/sim/engine/finder"

	"github.com/go-gl/mathgl/mgl32"
)

type RoadGrid struct {
	finder *finder.ElementFinder
	grid   *graph.Graph
}

func NewRoadGrid(finder *finder.ElementFinder) *RoadGrid {
	grid := RoadGrid{
		finder: finder,
		grid:   graph.NewGraph()}
	return &grid
}

// Adds a ;ome tp the road grid, returning the line ID, start node ID, and end node ID, in that order
func (p *RoadGrid) AddLine(start, end mgl32.Vec2, capacity int64, startNode, endNode int64) (int64, int64, int64) {
	line := RoadLine{capacity: capacity}

	if startNode == endNode && startNode != -1 {
		fmt.Printf("Roads must be between nodes and cannot (for a single line) loop\n")
		return -1, startNode, endNode
	} else if startNode != -1 && endNode != -1 {
		// This might be a duplicate line.
		connectionStatus := p.grid.AddConnection(startNode, endNode, &line)
		if connectionStatus.Status == graph.Exists {
			fmt.Printf("There already is a line from %v to %v.\n", startNode, endNode)
			return -1, startNode, endNode
		} else {
			mailroom.NewPowerLineChannel <- geometry.NewIdLine(connectionStatus.Id, [2]mgl32.Vec2{start, end})
		}
	}

	if startNode == -1 {
		startNode = p.grid.AddNode(&RoadTerminus{location: start})
		p.finder.AddElementChannel <- finder.NewElement(startNode, finder.RoadTerminus, []mgl32.Vec2{start})
	}

	if endNode == -1 {
		endNode = p.grid.AddNode(&RoadTerminus{location: end})
		p.finder.AddElementChannel <- finder.NewElement(endNode, finder.RoadTerminus, []mgl32.Vec2{end})
	}

	connectionStatus := p.grid.AddConnection(startNode, endNode, &line)
	mailroom.NewRoadLineChannel <- geometry.NewIdLine(connectionStatus.Id, [2]mgl32.Vec2{start, end})

	return connectionStatus.Id, startNode, endNode
}
