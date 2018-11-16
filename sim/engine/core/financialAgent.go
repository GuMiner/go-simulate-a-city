package core

import (
	"fmt"
	"go-simulate-a-city/sim/config"
)

type FinancialAgent struct {
	savings            float32
	TransactionChannel chan Transaction
	ControlChannel     chan int
}

func NewFinancialAgent() FinancialAgent {
	agent := FinancialAgent{
		savings:            config.Config.Sim.StartingSavings,
		TransactionChannel: make(chan Transaction),
		ControlChannel:     make(chan int)}

	go agent.Run()
	return agent
}

func (f *FinancialAgent) Run() {
	for {
		select {
		case t := <-f.TransactionChannel:
			f.savings -= t.Amount
			if f.savings > -config.Config.Sim.MaxDebt {
				// TODO: send a message indicating you have lost the game.
			}
			fmt.Printf("> Purchased a %v for %.0f. Savings: %.0f\n", t.Name, t.Amount, f.savings)
			break
		case _ = <-f.ControlChannel:
			return
		}
	}
}
