package actionagent

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

func (agent *Agent) ListenSignal() {
	sigChan := make(chan os.Signal, 10)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGUSR1)
	for {
		select {
		case _sig := <-sigChan:
			sig := _sig.(syscall.Signal)
			logrus.Printf("received a signal: %s (%d)\n", sig, sig)

			switch sig {
			case syscall.SIGTERM:
				agent.doSignal(sig)
				agent.Teardown(int(sig))

			case syscall.SIGUSR1:
				logrus.Println("nothing")

			default:
				agent.doSignal(sig)
			}
		}
	}
}

// doSignal handle signal:
// if script script exist, invoke script;
// otherwise, pass signal to script run.
func (agent *Agent) doSignal(sig syscall.Signal) {
	sigScript := getSigScriptPath(sig)
	_, err := os.Stat(sigScript)
	if err == nil {
		// script exist
		sigtermCmd := exec.Command(sigScript)
		sigtermCmd.Stdout = os.Stdout
		sigtermCmd.Stderr = os.Stderr
		err = sigtermCmd.Run()
		if err != nil {
			logrus.Println(err)
		}
	} else {
		agent.passSignalToRun(sig)
	}
}

// passSignalToRun pass signal to `run` directly.
func (agent *Agent) passSignalToRun(sig syscall.Signal) {
	if agent.EasyUse.RunProcess != nil {
		logrus.Printf("pass signal: %s (%d) to action run directly", sig, sig)
		err := agent.EasyUse.RunProcess.Signal(sig)
		if err != nil {
			logrus.Println(err)
		}
	}
}

// SIGTERM -> when_sig_15
// SIGINT  -> when_sig_2
func getSigScriptPath(sig syscall.Signal) string {
	return fmt.Sprintf("/opt/action/when_sig_%d", int(sig))
}
