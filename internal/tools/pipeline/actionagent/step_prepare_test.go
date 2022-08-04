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
			commands: []string{"sleep 120", "cat /etc/hosts", "echo 'hello world'"},
			want: `sleep 120
cat /etc/hosts
echo 'hello world'
`,
		},
		{
			name: "string commands",
			commands: `npm i @terminus/trnw-cli -g
trnw-cli build apk -T -A -t assembleRelease \
    --proxy \
    --env-default mall-((PROJECT_TYPE))/.env.default`,
			want: `npm i @terminus/trnw-cli -g
trnw-cli build apk -T -A -t assembleRelease \
    --proxy \
    --env-default mall-((PROJECT_TYPE))/.env.default
`,
		},
		{
			name: "interface commands",
			commands: []interface{}{
				"sleep 120",
				"tail",
				1,
			},
			want: `sleep 120
tail
1
`,
		},
		{
			name:     "invalid commands",
			commands: 123,
			want: `123
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
		agent.setupCommandScript()
		if tt.got != tt.want {
			t.Errorf("%q. setupCommandScript() = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}
