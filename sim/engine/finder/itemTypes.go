package finder

// Defines the types of object this finder can find.

type ItemType int

const (
	PowerLine ItemType = iota
	PowerPlant
	RoadLine
)
