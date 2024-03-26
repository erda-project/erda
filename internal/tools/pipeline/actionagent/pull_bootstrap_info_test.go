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
	"github.com/stretchr/testify/assert"
)

func Test_replaceEnvExpr(t *testing.T) {
	type args struct {
		privateEnvs map[string]string
	}
	envs := map[string]string{
		"DICE_MEMORY":    "5MB",
		"DICE_CORE":      "1",
		"DICE_WORKSPACE": "dev",
	}
	patch := monkey.Patch(os.Getenv, func(s string) string {
		return envs[s]
	})
	defer patch.Unpatch()
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "parse env_expr from env",
			args: args{
				privateEnvs: map[string]string{
					"ACTION_MEMORY": "${{ envs.DICE_MEMORY }}",
					"ACTION_CORE":   "${{ envs.DICE_CORE }}",
				},
			},
			want: map[string]string{
				"ACTION_MEMORY": "5MB",
				"ACTION_CORE":   "1",
			},
		},
		{
			name: "parse env from env and map",
			args: args{
				privateEnvs: map[string]string{
					"ACTION_MEMORY": "${{ envs.DICE_MEMORY }}",
					"ACTION_CORE":   "${{ envs.DICE_CORE }}",
					"ACTION_PARAM":  " it is ${{ envs.DICE_PARAM }}",
					"DICE_PARAM":    "dice_param",
				},
			},
			want: map[string]string{
				"ACTION_MEMORY": "5MB",
				"ACTION_CORE":   "1",
				"ACTION_PARAM":  " it is dice_param",
				"DICE_PARAM":    "dice_param",
			},
		},
		{
			name: "env_expr loop call",
			args: args{
				privateEnvs: map[string]string{
					"ACTION_MEMORY": "${{ envs.ACTION_CORE }}",
					"ACTION_CORE":   "${{ envs.ACTION_MEMORY }}",
				},
			},
			want: map[string]string{
				"ACTION_MEMORY": "${{ envs.ACTION_CORE }}",
				"ACTION_CORE":   "${{ envs.ACTION_MEMORY }}",
			},
		},
		{
			name: "env_expr recursive call",
			args: args{
				privateEnvs: map[string]string{
					"ACTION_MEMORY":    "${{ envs.ACTION_CORE }}",
					"ACTION_CORE":      "${{ envs.ACTION_WORKSPACE }}",
					"ACTION_WORKSPACE": "${{ envs.DICE_MEMORY }}",
				},
			},
			want: map[string]string{
				"ACTION_MEMORY":    "5MB",
				"ACTION_CORE":      "5MB",
				"ACTION_WORKSPACE": "5MB",
			},
		},
		{
			name: "env_Expr in non-action prefix",
			args: args{
				privateEnvs: map[string]string{
					"ACTION_PARAM": " it is ${{ envs.DICE_PARAM }}",
					"DICE_PARAM":   "dice_param",
					"DICE_ENV":     "${{ envs.DICE_PARAM }}",
					"DICE_A":       "${{ envs.ACTION_MEMORY }}",
					"ACTION_A":     "${{ envs.DICE_A }}",
				},
			},
			want: map[string]string{
				"ACTION_PARAM": " it is dice_param",
				"DICE_PARAM":   "dice_param",
				"DICE_ENV":     "${{ envs.DICE_PARAM }}",
				"DICE_A":       "${{ envs.ACTION_MEMORY }}",
				"ACTION_A":     "${{ envs.ACTION_MEMORY }}",
			},
		},
		{
			name: "",
			args: args{
				privateEnvs: map[string]string{
					"ACTION_COMMAND": "go build -o ${{ envs.DICE_WORKSPACE }} main.go",
				},
			},
			want: map[string]string{
				"ACTION_COMMAND": "go build -o dev main.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				Arg: &AgentArg{
					PrivateEnvs: tt.args.privateEnvs,
				},
			}
			agent.replaceEnvExpr()
			assert.Equal(t, tt.want, agent.Arg.PrivateEnvs)
		})
	}
}

func TestCommandReplaceEnv(t *testing.T) {
	type args struct {
		commands    interface{}
		privateEnvs map[string]string
	}

	envs := map[string]string{
		"DICE_MEMORY":    "5MB",
		"DICE_CORE":      "1",
		"DICE_WORKSPACE": "dev",
	}
	patch := monkey.Patch(os.Getenv, func(s string) string {
		return envs[s]
	})
	defer patch.Unpatch()

	testcase := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "test string from env and privateEnvs is nil",
			args: args{
				commands: "echo ${{ envs.DICE_WORKSPACE }}",
			},
			want: []string{"echo dev"},
		},
		{
			name: "test string from privateEnvs",
			args: args{
				privateEnvs: map[string]string{
					"ACTION_COMMAND": "go build -o ${{ envs.DICE_WORKSPACE }} main",
				},
				commands: "echo ${{ envs.ACTION_COMMAND }}",
			},
			want: []string{"echo go build -o dev main"},
		},
		{
			name: "test []string",
			args: args{
				privateEnvs: map[string]string{
					"ACTION_COMMAND": "go build -o ${{ envs.DICE_WORKSPACE }} main",
				},
				commands: []string{"echo ${{ envs.ACTION_COMMAND }}", "echo ${{ envs.DICE_CORE }}"},
			},
			want: []string{"echo go build -o dev main", "echo 1"},
		},
		{
			name: "test []interface{}",
			args: args{
				privateEnvs: map[string]string{
					"ACTION_SERVICE": "The memory is ${{ envs.DICE_MEMORY }}",
				},
				commands: []interface{}{
					"echo ${{ envs.ACTION_SERVICE }}",
					"import ${{ envs.DICE_CORE }}",
				},
			},
			want: []string{"echo The memory is 5MB", "import 1"},
		},
	}

	for _, tt := range testcase {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				Arg: &AgentArg{
					PrivateEnvs: tt.args.privateEnvs,
					Commands:    tt.args.commands,
				},
			}

			agent.replaceEnvExpr()
			assert.Equal(t, tt.want, agent.Arg.Commands)
		})
	}
}
