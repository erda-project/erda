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

package taskop

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func Test_prepare_generateOpenapiTokenForPullBootstrapInfo(t *testing.T) {
	type args struct {
		task *spec.PipelineTask
	}
	tests := []struct {
		name    string
		pre     prepare
		args    args
		wantErr bool
	}{
		{
			name: "test_api-test",
			pre:  prepare{},
			args: args{
				task: &spec.PipelineTask{
					Type: apistructs.ActionTypeAPITest,
				},
			},
			wantErr: false,
		},
		{
			name: "test_wait",
			pre:  prepare{},
			args: args{
				task: &spec.PipelineTask{
					Type: apistructs.ActionTypeWait,
				},
			},
			wantErr: false,
		},
		{
			name: "test_snippet",
			pre:  prepare{},
			args: args{
				task: &spec.PipelineTask{
					Type: apistructs.ActionTypeSnippet,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.pre.generateOpenapiTokenForPullBootstrapInfo(tt.args.task); (err != nil) != tt.wantErr {
				t.Errorf("generateOpenapiTokenForPullBootstrapInfo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_condition(t *testing.T) {
	task := &spec.PipelineTask{Extra: spec.PipelineTaskExtra{Action: pipelineyml.Action{If: "${{   1 == 1   }}"}}}
	b := condition(task)
	assert.Equal(t, false, b)
}
