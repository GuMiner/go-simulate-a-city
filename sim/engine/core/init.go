package core

import (
	"go-simulate-a-city/sim/core/mailroom"
	"go-simulate-a-city/sim/engine/core/agent"
)

var CoreTimer agent.Timer
var CoreFinances agent.FinancialAgent

func Init() {
	CoreTimer = agent.NewTimer()
	mailroom.CoreTimerRegChannel = CoreTimer.RegistrationChannel

	CoreFinances = agent.NewFinancialAgent()
}
