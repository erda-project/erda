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

	"bou.ke/monkey"

	"github.com/erda-project/erda/pkg/filehelper"
)

func Test_setupScript(t *testing.T) {
	tests := []struct {
		name     string
		commands interface{}
		got      string
		want     string
	}{
		{
			name:     "string slice commands",
			commands: []string{"sleep 120"},
			want: `#!/bin/sh
set -e

echo + "sleep 120"
sleep 120 || ((echo "- FAIL! exit code: $?") && false)
echo

`,
		},
		{
			name: "string commands",
			commands: `npm i @terminus/trnw-cli -g
            trnw-cli build apk -T -A -t assembleRelease \
                        --proxy \
                        --env-default mall-((PROJECT_TYPE))/.env.default`,
			want: `#!/bin/sh
set -e

echo + "npm i @terminus/trnw-cli -g\n            trnw-cli build apk -T -A -t assembleRelease \\\n                        --proxy \\\n                        --env-default mall-((PROJECT_TYPE))/.env.default"
npm i @terminus/trnw-cli -g
            trnw-cli build apk -T -A -t assembleRelease \
                        --proxy \
                        --env-default mall-((PROJECT_TYPE))/.env.default || ((echo "- FAIL! exit code: $?") && false)
echo

`,
		},
		{
			name: "interface commands",
			commands: []interface{}{
				"sleep 120",
				"tail",
				1,
			},
			want: `#!/bin/sh
set -e

echo + "sleep 120"
sleep 120 || ((echo "- FAIL! exit code: $?") && false)
echo

echo + "tail"
tail || ((echo "- FAIL! exit code: $?") && false)
echo

echo + "1"
1 || ((echo "- FAIL! exit code: $?") && false)
echo

`,
		},
		{
			name:     "invalid commands",
			commands: 123,
			want: `#!/bin/sh
set -e

echo + "123"
123 || ((echo "- FAIL! exit code: $?") && false)
echo

`,
		},
	}
	for _, tt := range tests {
		agent := &Agent{
			Arg: &AgentArg{
				Commands: tt.commands,
			},
		}
		monkey.Patch(filehelper.CreateFile, func(absPath string, content string, perm os.FileMode) error {
			tt.got = content
			return nil
		})
		agent.setupScript()
		if tt.got != tt.want {
			t.Errorf("%q. setupScript() = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}
