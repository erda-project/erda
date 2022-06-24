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
			if got := GenEnvKeyWithPrefix(tt.envPrefix, tt.key); got != tt.want {
				t.Errorf("GenEnvKeyWithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
