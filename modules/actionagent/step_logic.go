package actionagent

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func (agent *Agent) logic() {

	// go to ${WORKDIR}
	if err := os.Chdir(agent.EasyUse.ContainerWd); err != nil {
		agent.AppendError(err)
		return
	}

	actionRun := exec.Command(agent.EasyUse.RunScript)
	actionRun.Stdout = io.MultiWriter(os.Stdout, agent.EasyUse.RunMultiStdout)
	actionRun.Stderr = io.MultiWriter(os.Stderr, agent.EasyUse.RunMultiStderr)
	if err := actionRun.Start(); err != nil {
		agent.AppendError(err)
		return
	}
	agent.EasyUse.RunProcess = actionRun.Process
	if err := actionRun.Wait(); err != nil {
		agent.ExitCode = 1
		if err.Error() != fmt.Sprintf(`command "%s": exit status 1`, agent.EasyUse.RunScript) {
			logrus.Println(err)
			agent.AppendError(err)
		}
	}
}
