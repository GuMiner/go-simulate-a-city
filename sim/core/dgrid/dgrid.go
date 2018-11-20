package dgrid

// Defines a distributed non-directed grid defined by channel communication
type DistributedGrid struct {
	nodes map[int]*GridNode
	edges map[int]*GridConnection

	newNodeIndex       int
	newConnectionIndex int
}

func NewDistributedGrid() *DistributedGrid {
	distributedGrid := DistributedGrid{
		nodes:              make(map[int]*GridNode),
		edges:              make(map[int]*GridConnection),
		newNodeIndex:       0,
		newConnectionIndex: 0}

	go distributedGrid.run()
	return &distributedGrid
}

func (d *DistributedGrid) run() {
	for {
		// TODO: I need to figure out how to implement this properly.
	}
}

func (d *DistributedGrid) AddNode(data interface{}) int {
	return 1
}
