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

type ConnectionResult struct {
	Status ConnectionStatus
	Id     int64
}

// Defines a thread-safe bi-directional graph data structure, storing arbitrary node / connection data
// TODO: Implement edit methods and data retrieval methods
type Graph struct {
	nodes              map[int64]*Node
	connections        map[int64]*Connection
	newNodeIndex       int64
	newConnectionIndex int64

	nodesLock sync.Mutex

	connectionEditBuffer chan ConnectionEdit
	nodeEditBuffer       chan NodeEdit

	connectionEditRegistrations []chan ConnectionEdit
	nodeEditRegistrations       []chan NodeEdit
	ConnectionEditRegChannel    chan chan ConnectionEdit
	NodeEditRegChannel          chan chan NodeEdit
	ControlChannel              chan int64
}

func NewGraph() *Graph {
	graph := Graph{
		nodes:                       make(map[int64]*Node),
		connections:                 make(map[int64]*Connection),
		newNodeIndex:                0,
		newConnectionIndex:          0,
		connectionEditBuffer:        make(chan ConnectionEdit, 10),
		nodeEditBuffer:              make(chan NodeEdit, 10),
		connectionEditRegistrations: make([]chan ConnectionEdit, 0),
		nodeEditRegistrations:       make([]chan NodeEdit, 0),
		ConnectionEditRegChannel:    make(chan chan ConnectionEdit),
		NodeEditRegChannel:          make(chan chan NodeEdit),
		ControlChannel:              make(chan int64)}

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
func (d *Graph) AddConnection(first, second int64, data interface{}) ConnectionResult {
	d.nodesLock.Lock()
	defer d.nodesLock.Unlock()

	if _, ok := d.nodes[first]; ok {
		if _, ok = d.nodes[second]; ok {
			if _, ok = d.nodes[first].connections[second]; ok {
				return ConnectionResult{Status: Exists, Id: d.nodes[first].connections[second].Id}
			} else {
				connectionIdx := d.newConnectionIndex
				d.connections[connectionIdx] = &Connection{First: first, Second: second}
				d.nodes[first].connections[second] = nodeInternalConnection{Data: data, Id: connectionIdx}
				d.nodes[second].connections[first] = nodeInternalConnection{Data: data, Id: connectionIdx}

				d.connectionEditBuffer <- NewConnectionEdit(
					Add,
					d.nodes[first].data,
					connectionIdx,
					first,
					second,
					data)

				d.newConnectionIndex++
				return ConnectionResult{Status: Success, Id: connectionIdx}
			}
		}
	}

	return ConnectionResult{Status: NodesMissing, Id: -1}
}

// Deletes a connection, returning true if deleted, false if it was already deleted
func (d *Graph) DeleteConnection(first, second int64) bool {
	d.nodesLock.Lock()
	defer d.nodesLock.Unlock()

	if _, ok := d.nodes[first]; ok {
		if _, ok = d.nodes[second]; ok {
			if _, ok = d.nodes[first].connections[second]; ok {
				connectionId := d.nodes[first].connections[second].Id
				d.connectionEditBuffer <- NewConnectionEdit(
					Delete,
					d.nodes[first].data,
					connectionId,
					first,
					second,
					d.nodes[first].connections[second])
				delete(d.connections, connectionId)
				delete(d.nodes[first].connections, second)
				delete(d.nodes[second].connections, first)
				return true
			}
		}
	}

	return false
}

// Deletes a node, returning the node index if deleted, -1 if it is already gone.
func (d *Graph) DeleteNode(nodeIdx int64) int64 {
	d.nodesLock.Lock()
	defer d.nodesLock.Unlock()

	if _, ok := d.nodes[nodeIdx]; ok {
		for destinationNode, connectionData := range d.nodes[nodeIdx].connections {
			delete(d.connections, connectionData.Id)
			delete(d.nodes[destinationNode].connections, nodeIdx)
		}

		d.nodeEditBuffer <- NewNodeEdit(Delete, d.nodes[nodeIdx].data, nodeIdx)
		delete(d.nodes, nodeIdx)
		return nodeIdx
	}

	return -1
}

// Adds a new node, returning the node's ID
func (d *Graph) AddNode(data interface{}) int64 {
	d.nodesLock.Lock()
	defer d.nodesLock.Unlock()

	nodeIdx := d.newNodeIndex
	d.newNodeIndex++

	d.nodes[nodeIdx] = NewNode(data)
	d.nodeEditBuffer <- NewNodeEdit(Add, d.nodes[nodeIdx].data, nodeIdx)
	return nodeIdx
}

func (d *Graph) GetNode(nodeId int64) interface{} {
	d.nodesLock.Lock()
	defer d.nodesLock.Unlock()

	node := d.nodes[nodeId]
	if node == nil {
		return nil
	}

	return node.data
}

func (d *Graph) GetConnection(connectionId int64) interface{} {
	d.nodesLock.Lock()
	defer d.nodesLock.Unlock()

	connection := d.connections[connectionId]
	if connection == nil {
		return nil
	}

	return d.nodes[connection.First].connections[connection.Second].Data
}
