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
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/internal/tools/pipeline/actionagent/filewatch"
)

var (
	endFlagForBreakpoint = "EOF"
)

// Teardown teardown agent, including prestop, callback.
func (agent *Agent) Teardown(exitCode ...int) {
	if len(exitCode) > 0 {
		agent.ExitCode = exitCode[0]
	}
	agent.PreStop()
	agent.stop()
	agent.Callback()
}

// Exit exit agent.
func (agent *Agent) Exit() {
	os.Exit(agent.ExitCode)
}

func (agent *Agent) writeEndFlagLine() {
	if agent.EasyUse.RunMultiStdout != nil {
		if _, err := fmt.Fprintf(agent.EasyUse.RunMultiStdout, "\n%s\n", agent.EasyUse.FlagEndLineForTail); err != nil {
			logrus.Println("stdout append flag err:", err)
		}
	}
	if agent.EasyUse.RunMultiStderr != nil {
		if _, err := fmt.Fprintf(agent.EasyUse.RunMultiStderr, "\n%s\n", agent.EasyUse.FlagEndLineForTail); err != nil {
			logrus.Println("stderr append flag err:", err)
		}
	}
}

// CheckForBreakpointOnFailure if step up breakpoint on failure
// waiting breakpointExitPostFile to be written
func (agent *Agent) CheckForBreakpointOnFailure(breakPointFile string) {
	if !agent.Arg.DebugOnFailure {
		return
	}
	logrus.Infof("starting debug on failure, waiting for user's decision:")
	logrus.Infof("1) let the task success, use cmd: echo 0 > %s", breakPointFile)
	logrus.Infof("2) let the task fail, use cmd: echo 1 > %s", breakPointFile)
	ctx := context.Background()
	if agent.Arg.DebugTimeout != nil {
		ctx, _ = context.WithTimeout(ctx, *agent.Arg.DebugTimeout)
	}

	breakpointWatcher, err := filewatch.New(ctx)
	breakpointWatcher.EndLineForTail = endFlagForBreakpoint
	defer breakpointWatcher.Close()
	if err != nil {
		logrus.Errorf("failed to create breakpoint watcher, err: %v", err)
		return
	}
	breakpointWatcher.RegisterTailHandler(breakPointFile, func(line string, allLines []string) error {
		strExitCode := strings.TrimSuffix(line, "\n")
		logrus.Infof("Breakpoint exiting with exit code " + strExitCode)
		defer func() {
			if err := agent.writeEndFlagForBreakpoint(breakPointFile); err != nil {
				logrus.Errorf("failed to write end flag for breakpoint, err: %v", err)
			}
		}()
		exitCode, err := strconv.Atoi(strExitCode)
		if err != nil {
			logrus.Errorf("error occurred while reading breakpoint exit code, err: %v", err)
			return err
		}
		agent.ExitCode = exitCode
		return nil
	})
}

func (agent *Agent) writeEndFlagForBreakpoint(breakPointFile string) error {
	writer, err := os.OpenFile(breakPointFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer writer.Close()
	_, err = writer.WriteString(fmt.Sprintf("%s\n", endFlagForBreakpoint))
	if err != nil {
		return err
	}
	return nil
}
