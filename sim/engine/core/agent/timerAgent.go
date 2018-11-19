package agent

import (
	"go-simulate-a-city/sim/engine/core/dto"
	"time"
)

// Defines a high-level timer suitable for UI updates
//  and a low-level timer suitable for simulation updates
type Timer struct {
	registeredChannels []chan dto.Time
	highResChannels    []chan time.Time

	ControlChannel      chan int
	RegistrationChannel chan chan dto.Time
	HighResRegChannel   chan chan time.Time
}

func NewTimer() Timer {
	timer := Timer{
		registeredChannels:  make([]chan dto.Time, 0),
		highResChannels:     make([]chan time.Time, 0),
		ControlChannel:      make(chan int),
		RegistrationChannel: make(chan chan dto.Time),
		HighResRegChannel:   make(chan chan time.Time)}

	ticker := time.NewTicker(100 * time.Millisecond)
	highResTicker := time.NewTicker(20 * time.Millisecond)
	go timer.run(ticker)
	go timer.runHighRes(highResTicker)

	return timer
}

func (t *Timer) runHighRes(ticker *time.Ticker) {
	for {
		time := <-ticker.C
		for _, channel := range t.highResChannels {
			channel <- time
		}

		select {
		case registration := <-t.HighResRegChannel:
			t.highResChannels = append(t.highResChannels, registration)
			break
		case _ = <-t.ControlChannel:
			ticker.Stop()
			return
		default:
			break
		}
	}
}

func (t *Timer) run(ticker *time.Ticker) {
	time := dto.NewTime()

	for {
		// Blocks till we advance 100 ms and updates our time.
		_ = <-ticker.C
		time.Update(0.1)

		// Send the time to everyone!
		for _, channel := range t.registeredChannels {
			channel <- time
		}

		select {
		case registration := <-t.RegistrationChannel:
			t.registeredChannels = append(t.registeredChannels, registration)
			break
		case _ = <-t.ControlChannel:
			ticker.Stop()
			return
		default:
			break
		}
	}
}