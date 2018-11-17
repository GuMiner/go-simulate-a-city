package agent

import (
	"fmt"
	"go-simulate-a-city/sim/config"
	"go-simulate-a-city/sim/engine/core/dto"
)

type FinancialAgent struct {
	savings            float32
	TransactionChannel chan dto.Transaction
	ControlChannel     chan int
}

func NewFinancialAgent() FinancialAgent {
	agent := FinancialAgent{
		savings:            config.Config.Sim.StartingSavings,
		TransactionChannel: make(chan dto.Transaction),
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
