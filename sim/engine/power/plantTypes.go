package power

import (
	"sim/config"
	"sim/core/dto/editorengdto"
)

type PowerPlantSize int

const (
	Small PowerPlantSize = iota
	Large
)

func GetPowerOutputAndSize(plantType string, plantSize PowerPlantSize) (output int, size int) {
	plant := config.Config.Power.PowerPlantTypes[plantType]

	switch plantSize {
	case Small:
		output = plant.SmallOutput
		size = plant.SmallSize
	default: // Large
		output = plant.LargeOutput
		size = plant.LargeSize
	}

	return output, size
}

func GetPlantType(itemSelection editorengdto.ItemSubSelection) string {
	return config.Config.Power.IdToNameMap[int(itemSelection)]
}

func GetPlantCost(plantType string) float32 {
	plant := config.Config.Power.PowerPlantTypes[plantType]
	return plant.Cost
}
