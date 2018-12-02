package road

import (
	"fmt"
	"go-simulate-a-city/sim/core/dto/geometry"
	"go-simulate-a-city/sim/core/graph"
	"go-simulate-a-city/sim/core/mailroom"

	"github.com/go-gl/mathgl/mgl32"
)

type RoadGrid struct {
	grid *graph.Graph
}

func NewRoadGrid() *RoadGrid {
	grid := RoadGrid{
		grid: graph.NewGraph()}
	return &grid
}

func (p *RoadGrid) AddLine(start, end mgl32.Vec2, capacity int64, startNode, endNode int64) int64 {
	line := RoadLine{capacity: capacity}

	if startNode == endNode && startNode != -1 {
		fmt.Printf("Roads must be between nodes and cannot (for a single line) loop\n")
		return -1
	} else if startNode != -1 && endNode != -1 {
		// This might be a duplicate line.
		connectionStatus := p.grid.AddConnection(startNode, endNode, &line)
		if connectionStatus.Status == graph.Exists {
			fmt.Printf("There already is a line from %v to %v.\n", startNode, endNode)
			return -1
		} else {
			mailroom.NewPowerLineChannel <- geometry.NewIdLine(connectionStatus.Id, [2]mgl32.Vec2{start, end})
		}
	}

	if startNode == -1 {
		startNode = p.grid.AddNode(&RoadTerminus{location: start})
	}

	if endNode == -1 {
		endNode = p.grid.AddNode(&RoadTerminus{location: end})
	}

	connectionStatus := p.grid.AddConnection(startNode, endNode, &line)
	mailroom.NewRoadLineChannel <- geometry.NewIdLine(connectionStatus.Id, [2]mgl32.Vec2{start, end})

	return connectionStatus.Id
}
