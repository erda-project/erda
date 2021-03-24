package actionagent

import (
	"os"
)

// Teardown teardown agent, including prestop, callback.
func (agent *Agent) Teardown(exitCode ...int) {
	if len(exitCode) > 0 {
		agent.ExitCode = exitCode[0]
	}
	agent.PreStop()
	agent.Callback()
}

// Exit exit agent.
func (agent *Agent) Exit() {
	os.Exit(agent.ExitCode)
}
