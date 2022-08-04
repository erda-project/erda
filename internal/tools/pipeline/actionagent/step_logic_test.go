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
	"os"
	"testing"
)

func Test_runCommands(t *testing.T) {
	agent := &Agent{
		EasyUse: EasyUse{
			ContainerWd:    "./",
			RunMultiStdout: os.Stdout,
			RunMultiStderr: os.Stderr,
			CommandScript:  "./command",
		},
		Arg: &AgentArg{
			Shell: "sh",
		},
	}
	agent.runCommands()
}
