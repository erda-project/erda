// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
