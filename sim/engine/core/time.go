package core

import (
	"fmt"
	"go-simulate-a-city/sim/config"
)

type Time struct {
	SimTime float32
	DayTime float32
	Days    int
}

func newTime() Time {
	return Time{
		SimTime: 0,
		DayTime: 0,
		Days:    0}
}

func (t *Time) update(secondsDelta float32) {
	t.SimTime += 0.1
	t.DayTime += 0.1
	if t.DayTime > config.Config.Sim.SecondsPerDay {
		t.DayTime -= config.Config.Sim.SecondsPerDay
		t.Days++

		fmt.Printf(">>> Advanced to day %v <<<\n", t.Days)
	}
}
