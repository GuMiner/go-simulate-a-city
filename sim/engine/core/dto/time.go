package dto

import (
	"fmt"
	"sim/config"
)

// Defines a simple time DTO
type Time struct {
	SimTime float32
	DayTime float32
	Days    int
}

func NewTime() Time {
	return Time{
		SimTime: 0,
		DayTime: 0,
		Days:    0}
}

func (t *Time) Update(secondsDelta float32) {
	t.SimTime += 0.1
	t.DayTime += 0.1
	if t.DayTime > config.Config.Sim.SecondsPerDay {
		t.DayTime -= config.Config.Sim.SecondsPerDay
		t.Days++

		fmt.Printf(">>> Advanced to day %v <<<\n", t.Days)
	}
}
