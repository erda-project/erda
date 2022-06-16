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

package env

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/tools/pipeline/actionagent"
)

func TestGenPipelineEnvKey(t *testing.T) {
	type arg struct {
		name      string
		key       string
		envPrefix string
		want      string
	}
	tests := []arg{
		{
			name:      "action param",
			key:       "command",
			envPrefix: actionagent.EnvActionParamPrefix,
			want:      "ACTION_COMMAND",
		},
		{
			name:      "pipeline secret",
			key:       "encrypted_password",
			envPrefix: EnvPipelineSecretPrefix,
			want:      "PIPELINE_SECRET_ENCRYPTED_PASSWORD",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenPipelineEnvKey(tt.envPrefix, tt.key); got != tt.want {
				t.Errorf("GenPipelineEnvKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTaskSourceParamsFromEnv(t *testing.T) {
	envs := map[string]string{
		"ACTION_COMMAND":             "go build -o web-server main.go",
		"ACTION_CONTEXT":             "/.pipeline/container/context/git-checkout",
		"ACTION_SERVICE":             "web-server",
		"ACTION_TARGET":              "web-server",
		"BCC":                        "cc",
		"BP_REPO_DEFAULT_VERSION":    "",
		"BP_REPO_PREFIX":             "",
		"CONTEXTDIR":                 "/.pipeline/container/context",
		"DATE_YYYYMMDD":              "20220615",
		"DICE_APPLICATION_ID":        "10",
		"ENCRCYPT_KEYQ":              "",
		"GITTAR_AUTHOR":              "erda",
		"METAFILE":                   "/.pipeline/container/metadata/go-demo/metadata",
		"PIPELINE_CRON_EXPR":         "",
		"PIPELINE_CRON_TRIGGER_TIME": "",
		"PIPELINE_ID":                "129771609784449",
		"PIPELINE_SECRET_ABCB":       "bbb",
		"PIPELINE_SECRET_BCC":        "cc",
		"PIPELINE_TYPE":              "",
		"QWCE":                       "eqw",
		"QWEEC":                      "",
		"UPLOADDIR":                  "/.pipeline/container/uploaddir",
		"WORKDIR":                    "/.pipeline/container/context/go-demo",
	}
	params := GetTaskSourceParamsFromEnv(actionagent.EnvActionParamPrefix, envs)
	t.Log(params)
	assert.Equal(t, "go build -o web-server main.go", params["command"])
}
