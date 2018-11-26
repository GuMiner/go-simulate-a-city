package graph

import (
	"sync"
)

type ConnectionStatus int

const (
	Success ConnectionStatus = iota
	Exists
	NodesMissing
)

// Defines a thread-safe bi-directional graph data structure, storing arbitrary node / connection data
// TODO: Implement edit methods and data retrieval methods
type Graph struct {
	nodes        map[int]*Node
	newNodeIndex int

	nodesLock sync.Mutex

	connectionEditBuffer chan ConnectionEdit
	nodeEditBuffer       chan NodeEdit

	connectionEditRegistrations []chan ConnectionEdit
	nodeEditRegistrations       []chan NodeEdit
	ConnectionEditRegChannel    chan chan ConnectionEdit
	NodeEditRegChannel          chan chan NodeEdit
	ControlChannel              chan int
}

func NewGraph() *Graph {
	graph := Graph{
		nodes:                       make(map[int]*Node),
		newNodeIndex:                0,
		connectionEditBuffer:        make(chan ConnectionEdit, 10),
		nodeEditBuffer:              make(chan NodeEdit, 10),
		connectionEditRegistrations: make([]chan ConnectionEdit, 0),
		nodeEditRegistrations:       make([]chan NodeEdit, 0),
		ConnectionEditRegChannel:    make(chan chan ConnectionEdit),
		NodeEditRegChannel:          make(chan chan NodeEdit),
		ControlChannel:              make(chan int)}

	go graph.run()
	return &graph
}

func (d *Graph) run() {
	for {
		select {
		case reg := <-d.ConnectionEditRegChannel:
			d.connectionEditRegistrations = append(d.connectionEditRegistrations, reg)
		case reg := <-d.NodeEditRegChannel:
			d.nodeEditRegistrations = append(d.nodeEditRegistrations, reg)
		case connectionEdit := <-d.connectionEditBuffer:
			for _, reg := range d.connectionEditRegistrations {
				reg <- connectionEdit
			}
		case nodeEdit := <-d.nodeEditBuffer:
			for _, reg := range d.nodeEditRegistrations {
				reg <- nodeEdit
			}
		case _ = <-d.ControlChannel:
			return
		}
	}
}

// Adds a connection between two nodes, returning the status of the operation
func (d *Graph) AddConnection(first, second int, data interface{}) ConnectionStatus {
	d.nodesLock.Lock()
	defer d.nodesLock.Unlock()

	if _, ok := d.nodes[first]; ok {
		if _, ok = d.nodes[second]; ok {
			if _, ok = d.nodes[first].connections[second]; ok {
				return Exists
			} else {
				d.nodes[first].connections[second] = data
				d.nodes[second].connections[first] = data
				d.connectionEditBuffer <- NewConnectionEdit(
					Add,
					d.nodes[first].data,
					first,
					second,
					data)
				return Success
			}
		}
	}

	return NodesMissing
}

// Deletes a connection, returning true if deleted, false if it was already deleted
func (d *Graph) DeleteConnection(first, second int) bool {
	d.nodesLock.Lock()
	defer d.nodesLock.Unlock()

	if _, ok := d.nodes[first]; ok {
		if _, ok = d.nodes[second]; ok {
			if _, ok = d.nodes[first].connections[second]; ok {
				d.connectionEditBuffer <- NewConnectionEdit(
					Delete,
					d.nodes[first].data,
					first,
					second,
					d.nodes[first].connections[second])
				delete(d.nodes[first].connections, second)
				delete(d.nodes[second].connections, first)
				return true
			}
		}
	}

	return false
}

// Deletes a node, returning the node index if deleted, -1 if it is already gone.
func (d *Graph) DeleteNode(nodeIdx int) int {
	d.nodesLock.Lock()
	defer d.nodesLock.Unlock()

	if _, ok := d.nodes[nodeIdx]; ok {
		for destinationNode, _ := range d.nodes[nodeIdx].connections {
			delete(d.nodes[destinationNode].connections, nodeIdx)
		}

		d.nodeEditBuffer <- NewNodeEdit(Delete, d.nodes[nodeIdx].data, nodeIdx)
		delete(d.nodes, nodeIdx)
		return nodeIdx
	}

	return -1
}

// Adds a new node, returning the node's ID
func (d *Graph) AddNode(data interface{}) int {
	d.nodesLock.Lock()
	defer d.nodesLock.Unlock()

	nodeIdx := d.newNodeIndex
	d.newNodeIndex++

	d.nodes[nodeIdx] = NewNode(data)
	d.nodeEditBuffer <- NewNodeEdit(Add, d.nodes[nodeIdx].data, nodeIdx)
	return nodeIdx
}
