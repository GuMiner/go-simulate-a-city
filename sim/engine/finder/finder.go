package finder

import (
	"go-simulate-a-city/sim/core"
)

// Defines how to quickly add, remove, and find points on our gameboard.

type ElementFinder struct {
	// Maps each element by item type to its Ids, which are guaranteed unique
	elements map[ItemType]map[int64]Element

	AddElementChannel     chan Element
	KNearestSearchChannel chan KNearestNodesQuery
}

func NewElementFinder() *ElementFinder {
	finder := ElementFinder{
		elements:              make(map[ItemType]map[int64]Element),
		AddElementChannel:     make(chan Element),
		KNearestSearchChannel: make(chan KNearestNodesQuery)}

	go finder.run()

	return &finder
}

func (e *ElementFinder) run() {
	for {
		select {
		case newElement := <-e.AddElementChannel:
			if _, ok := e.elements[newElement.Type]; !ok {
				e.elements[newElement.Type] = make(map[int64]Element)
			}

			e.elements[newElement.Type][newElement.Id] = newElement
		case search := <-e.KNearestSearchChannel:
			search.Results <- e.KNearest(search)
		}
	}
}

// Returns the K-nearest elements, searching via nodes
func (e *ElementFinder) KNearest(search KNearestNodesQuery) []*NodeWithDistance {
	nodes := core.NewSortableArray(search.Count)

	for _, itemType := range search.Types {
		if _, ok := e.elements[itemType]; ok {
			for _, element := range e.elements[itemType] {
				for idx, node := range element.Nodes {
					distance := node.Sub(search.Pos).Len()
					nodes.Add(NewNodeWithDistance(element.Id, itemType, node, idx, distance))
				}
			}
		}
	}

	resultSet := make([]*NodeWithDistance, len(nodes.Items))
	for _, item := range nodes.Items {
		resultSet = append(resultSet, item.(*NodeWithDistance))
	}
	return resultSet
}
