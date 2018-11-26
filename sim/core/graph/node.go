package graph

type Node struct {
	connections map[int]interface{}
	data        interface{}
}

func NewNode(data interface{}) *Node {
	return &Node{
		connections: make(map[int]interface{}, 0),
		data:        data}
}
