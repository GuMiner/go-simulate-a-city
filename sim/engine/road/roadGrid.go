package road

import (
	"fmt"
	"go-simulate-a-city/sim/core/dto/geometry"
	"go-simulate-a-city/sim/core/graph"
	"go-simulate-a-city/sim/core/mailroom"
	"go-simulate-a-city/sim/engine/finder"

	"github.com/go-gl/mathgl/mgl32"
)

func min64(a, b int64) int64 {
	if a > b {
		return b
	}
	return a
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

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

// Adds a ;ome tp the road grid, returning the start node ID, line ID, and end node ID, in that order
func (p *RoadGrid) AddLine(start, end mgl32.Vec2, capacity int64, startNode, endNode int64) (int64, int64, int64) {
	line := NewRoadLine(capacity)

	if startNode == endNode && startNode != -1 {
		fmt.Printf("Roads must be between nodes and cannot (for a single line) loop\n")
		return -1, -1, -1
	} else if startNode != -1 && endNode != -1 {
		// This might be a duplicate line.
		connectionStatus := p.grid.AddConnection(startNode, endNode, line)
		if connectionStatus.Status == graph.Exists {
			fmt.Printf("There already is a line from %v to %v.\n", startNode, endNode)
			return -1, -1, -1
		} else {
			mailroom.NewRoadLineChannel <- geometry.NewIdLine(connectionStatus.Id, [2]mgl32.Vec2{start, end})

			go line.run()
			return startNode, connectionStatus.Id, endNode
		}
	}

	if startNode == -1 {
		terminus := NewRoadTerminus(start)
		//go terminus.run()

		startNode = p.grid.AddNode(terminus)
		terminus.Id = startNode
		p.finder.AddElementChannel <- finder.NewElement(startNode, finder.RoadTerminus, []mgl32.Vec2{start})
	}

	if endNode == -1 {
		terminus := NewRoadTerminus(end)
		//go terminus.run()
		endNode = p.grid.AddNode(terminus)
		terminus.Id = endNode
		p.finder.AddElementChannel <- finder.NewElement(endNode, finder.RoadTerminus, []mgl32.Vec2{end})
	}

	connectionStatus := p.grid.AddConnection(startNode, endNode, line)
	mailroom.NewRoadLineChannel <- geometry.NewIdLine(connectionStatus.Id, [2]mgl32.Vec2{start, end})

	// Hookup nodes to termii. TODO, this should use the grid data structure
	startTerminus := p.grid.GetNode(startNode).(*RoadTerminus)
	startTerminus.LineAddVehicleChannels[connectionStatus.Id] = line.AddVehicleChannel
	endTerminus := p.grid.GetNode(endNode).(*RoadTerminus)
	endTerminus.LineAddVehicleChannels[connectionStatus.Id] = line.AddVehicleChannel

	line.lowTerminus = min64(startNode, endNode)
	line.highTerminus = max64(startNode, endNode)
	if startNode > endNode {
		line.lowTerminusAddChannel = endTerminus.AddVehicleChannel
		line.highTerminusAddChannel = startTerminus.AddVehicleChannel
		line.lowPos = endTerminus.location
		line.highPos = startTerminus.location
	} else {
		line.lowTerminusAddChannel = startTerminus.AddVehicleChannel
		line.highTerminusAddChannel = endTerminus.AddVehicleChannel
		line.lowPos = startTerminus.location
		line.highPos = endTerminus.location
	}

	//go line.run()
	return startNode, connectionStatus.Id, endNode
}
