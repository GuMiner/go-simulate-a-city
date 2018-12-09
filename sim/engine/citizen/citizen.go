package citizen

import "go-simulate-a-city/sim/core/cmap"

type Citizen struct {
	Age int // In days
}

type CitizenManager struct {
	citizens *cmap.Map
}

func NewCitizenManager() *CitizenManager {
	manager := &CitizenManager{
		citizens: cmap.NewMap()}

	return manager
}
