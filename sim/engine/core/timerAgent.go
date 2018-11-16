package core

import (
	"time"
)

type Timer struct {
	registeredChannels []chan Time

	ControlChannel      chan int
	RegistrationChannel chan chan Time
}

func NewTimer() Timer {
	timer := Timer{
		registeredChannels:  make([]chan Time, 0),
		ControlChannel:      make(chan int),
		RegistrationChannel: make(chan chan Time)}

	ticker := time.NewTicker(100 * time.Millisecond)
	go timer.run(ticker)

	return timer
}

func (t *Timer) run(ticker *time.Ticker) {
	time := newTime()

	for {
		// Blocks till we advance 100 ms and updates our time.
		_ = <-ticker.C
		time.update(0.1)

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
