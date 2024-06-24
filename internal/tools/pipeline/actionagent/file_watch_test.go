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
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSlice(t *testing.T) {
	f := func(flogs *[]string, line string) {
		*flogs = append(*flogs, line)
	}
	var logs = &[]string{}
	f(logs, "l1")
	fmt.Println(logs)
	f(logs, "l2")
	fmt.Println(logs)
}

func TestMatchStdErrLine(t *testing.T) {
	regexpList := []*regexp.Regexp{}
	regexpStrList := []string{"^[a-z]*can't*"}
	for i := range regexpStrList {
		reg, err := regexp.Compile(regexpStrList[i])
		assert.NoError(t, err)
		regexpList = append(regexpList, reg)
	}
	assert.False(t, matchStdErrLine(regexpList, "failed to open file test.text"))
	assert.True(t, matchStdErrLine(regexpList, "can't open '/jfoejfoijlkjlj/jflejwf.txt': No such file or directory"))
}

func TestDesensitizeLine(t *testing.T) {
	var txtBlackList = []string{"abcdefg", "kskdudh&"}
	tt := []struct {
		line, want string
	}{
		{"admin", "admin"},
		{"abcdefg", "******"},
		{"admin123", "admin123"},
		{"docker login -u abcdefg -p kskdudh&", "docker login -u ****** -p ******"},
	}
	for _, v := range tt {
		desensitizeLine(txtBlackList, &v.line)
		assert.Equal(t, v.want, v.line)
	}

}

func TestAgent_BadWatchFiles(t *testing.T) {
	badAgent := Agent{}
	badAgent.watchFiles()
	assert.True(t, len(badAgent.Errs) > 0)
}

func TestAgent_watchFiles(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	agent := Agent{
		Ctx:    ctx,
		Cancel: cancel,
	}
	agent.watchFiles()
	t.Logf("no error here")
}

func TestCheckForBreakpointOnFailure(t *testing.T) {
	type arg struct {
		exitCode int
		debug    bool
	}
	testCases := []struct {
		name     string
		arg      arg
		exitCode int
	}{
		{
			name: "exit code 0",
			arg: arg{
				exitCode: 0,
				debug:    false,
			},
			exitCode: 0,
		},
		{
			name: "exit code 1 with debug continue",
			arg: arg{
				exitCode: 1,
				debug:    true,
			},
			exitCode: 0,
		},
		{
			name: "exit code 1 with debug failed",
			arg: arg{
				exitCode: 1,
				debug:    true,
			},
			exitCode: 2,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			agent := Agent{
				ExitCode: tc.arg.exitCode,
				Arg: &AgentArg{
					DebugOnFailure: tc.arg.debug,
					DebugTimeout:   &[]time.Duration{2 * time.Second}[0],
				},
			}

			tmpFile, err := os.CreateTemp("", "breakpoint")
			assert.NoError(t, err)
			fileName := tmpFile.Name()
			defer os.Remove(fileName)
			go func(f *os.File, exitCode int) {
				time.Sleep(1 * time.Second)
				f.WriteString(fmt.Sprintf("%d\n", exitCode))
			}(tmpFile, tc.exitCode)
			agent.CheckForBreakpointOnFailure(fileName)
			assert.Equal(t, tc.exitCode, agent.ExitCode)
		})
	}
}

func TestWatchFilesAndStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	agent := Agent{
		Ctx:    ctx,
		Cancel: cancel,
	}
	metaFile, err := os.CreateTemp("", "meta")
	assert.NoError(t, err)
	defer os.Remove(metaFile.Name())
	stdoutFile, err := os.CreateTemp("", "stdout")
	assert.NoError(t, err)
	defer os.Remove(stdoutFile.Name())
	stderrFile, err := os.CreateTemp("", "stderr")
	assert.NoError(t, err)
	defer os.Remove(stderrFile.Name())
	agent.EasyUse.ContainerMetaFile = metaFile.Name()
	agent.EasyUse.RunMultiStdoutFilePath = stdoutFile.Name()
	agent.EasyUse.RunMultiStderrFilePath = stderrFile.Name()
	agent.EasyUse.RunMultiStdout = stdoutFile
	agent.EasyUse.RunMultiStderr = stderrFile
	agent.watchFiles()

	stdoutFile.WriteString("stdout")
	stderrFile.WriteString("stderr")
	agent.writeEndFlagLine()
	agent.stop()
	t.Logf("no error here")
}
