package graph

type nodeInternalConnection struct {
	Data interface{}
	Id   int64
}

type Node struct {
	connections map[int64]nodeInternalConnection
	data        interface{}
}

func NewNode(data interface{}) *Node {
	return &Node{
		connections: make(map[int64]nodeInternalConnection, 0),
		data:        data}
}

type Connection struct {
	First, Second int64
}
