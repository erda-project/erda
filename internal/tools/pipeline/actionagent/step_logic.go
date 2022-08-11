// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package actionagent

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/pkg/filehelper"
)

// logic execute action run and command run
func (agent *Agent) logic() {
	if err := filehelper.CheckExist(agent.EasyUse.RunScript, false); err == nil {
		agent.actionRun()
	}
	if agent.Arg.Commands != nil {
		agent.runCommands()
	}
}

func (agent *Agent) actionRun() {
	// go to ${WORKDIR}
	if err := os.Chdir(agent.EasyUse.ContainerWd); err != nil {
		agent.AppendError(err)
		return
	}

	actionRun := exec.Command(agent.EasyUse.RunScript)
	actionRun.Stdout = io.MultiWriter(agent.EasyUse.RunMultiStdout)
	actionRun.Stderr = io.MultiWriter(agent.EasyUse.RunMultiStderr)
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

func (agent *Agent) runCommands() {
	// go to ${WORKDIR}
	if err := os.Chdir(agent.EasyUse.ContainerWd); err != nil {
		agent.AppendError(err)
		return
	}

	args := make([]string, len(agent.Arg.ShellArgs))
	copy(args, agent.Arg.ShellArgs)
	args = append(args, agent.EasyUse.CommandScript)
	commandRun := exec.Command(agent.Arg.Shell, args...)
	commandRun.Stdout = io.MultiWriter(agent.EasyUse.RunMultiStdout)
	commandRun.Stderr = io.MultiWriter(agent.EasyUse.RunMultiStderr)
	if err := commandRun.Start(); err != nil {
		agent.AppendError(err)
		return
	}
	agent.EasyUse.RunProcess = commandRun.Process
	if err := commandRun.Wait(); err != nil {
		agent.ExitCode = agent.getExitCode(err)
		if err.Error() != fmt.Sprintf(`command "%s": exit status %d`, agent.EasyUse.CommandScript, agent.ExitCode) {
			logrus.Println(err)
			agent.AppendError(err)
		}
	}
}

func (agent *Agent) getExitCode(err error) int {
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}
	return 1
}
