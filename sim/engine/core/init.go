package core

import "go-simulate-a-city/sim/engine/core/agent"

var CoreTimer agent.Timer
var CoreFinances agent.FinancialAgent

func Init() {
	CoreTimer = agent.NewTimer()
	CoreFinances = agent.NewFinancialAgent()
}
