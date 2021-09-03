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

package definition

import (
	"fmt"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/xormplus/xorm"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func Test_definitionAdaptor(t *testing.T) {
	type args struct {
		pipelineDefinition *PipelineDefinitionProcess
		session            *xorm.Session
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test_adaptor",
			args: args{
				pipelineDefinition: &PipelineDefinitionProcess{
					ID:              1,
					PipelineSource:  "test",
					PipelineYmlName: "test",
					SnippetConfig: &apistructs.SnippetConfigOrder{
						Source: "test",
						Name:   "test",
						SnippetLabels: []apistructs.SnippetLabel{
							{
								Key:   "test",
								Value: "test",
							},
						},
					},
					PipelineYml:           "version: \"1.1\"\nstages: []",
					PipelineCreateRequest: &apistructs.PipelineCreateRequestV2{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RegisterDefinitionHandler(func(definition PipelineDefinitionProcess, yml pipelineyml.PipelineYml) error {
				assert.Equal(t, definition, *tt.args.pipelineDefinition)
				return fmt.Errorf("error")
			})
			if err := definitionAdaptor(tt.args.pipelineDefinition); (err != nil) != tt.wantErr {
				t.Errorf("definitionAdaptor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
