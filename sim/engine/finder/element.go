package finder

import (
	"github.com/go-gl/mathgl/mgl32"
)

// Defines findeable items
type Element struct {
	Id    int64
	Type  ItemType
	Nodes []mgl32.Vec2
}

func NewElement(id int64, itemType ItemType, nodes []mgl32.Vec2) Element {
	return Element{
		Id:    id,
		Type:  itemType,
		Nodes: nodes}
}

// Defines a query to return the KNearestNodes of the given input types
type KNearestNodesQuery struct {
	Pos     mgl32.Vec2
	Types   []ItemType
	Count   int
	Results chan []*NodeWithDistance
}

func NewKNNQuery(pos mgl32.Vec2, itemType ItemType, count int, results chan []*NodeWithDistance) KNearestNodesQuery {
	query := KNearestNodesQuery{
		Pos:     pos,
		Types:   []ItemType{itemType},
		Count:   count,
		Results: results}

	return query
}

// Defines a node a distance away
type NodeWithDistance struct {
	Id   int64
	Type ItemType

	Pos       mgl32.Vec2
	NodeIndex int

	Distance float32
}

func NewNodeWithDistance(id int64, itemType ItemType, pos mgl32.Vec2, idx int, distance float32) *NodeWithDistance {
	return &NodeWithDistance{
		Id:        id,
		Type:      itemType,
		Pos:       pos,
		NodeIndex: idx,
		Distance:  distance}
}

// Implement Sortable (core)
func (n *NodeWithDistance) GetDistance() float32 {
	return n.Distance
}
