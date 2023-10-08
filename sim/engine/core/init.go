package core

import (
	"sim/core/mailroom"
	"sim/engine/core/agent"
)

var CoreTimer agent.Timer
var CoreFinances agent.FinancialAgent

func Init() {
	CoreTimer = agent.NewTimer()
	mailroom.CoreTimerRegChannel = CoreTimer.RegistrationChannel

	CoreFinances = agent.NewFinancialAgent()
}
