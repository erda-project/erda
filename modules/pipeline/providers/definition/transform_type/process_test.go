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

package transform_type

import (
	"testing"

	"github.com/erda-project/erda/apistructs"
)

func TestPipelineDefinitionProcess_Validate(t *testing.T) {
	type fields struct {
		PipelineSource        apistructs.PipelineSource
		PipelineYmlName       string
		PipelineYml           string
		SnippetConfig         *apistructs.SnippetConfig
		PipelineCreateRequest *apistructs.PipelineCreateRequestV2
	}
	tests := []struct {
		name    string
		fields  *fields
		wantErr bool
	}{
		{
			name:    "test_empty",
			fields:  nil,
			wantErr: true,
		},
		{
			name: "test_not_find_source",
			fields: &fields{
				PipelineSource: "",
			},
			wantErr: true,
		},
		{
			name: "test_empty_PipelineYmlName",
			fields: &fields{
				PipelineSource:  apistructs.PipelineSourceAutoTest,
				PipelineYmlName: "",
			},
			wantErr: true,
		},
		{
			name: "test_empty_PipelineYml",
			fields: &fields{
				PipelineSource:  apistructs.PipelineSourceAutoTest,
				PipelineYmlName: "test",
				PipelineYml:     "",
			},
			wantErr: true,
		},
		{
			name: "test_empty_PipelineCreateRequest",
			fields: &fields{
				PipelineSource:        apistructs.PipelineSourceAutoTest,
				PipelineYmlName:       "test",
				PipelineYml:           "test",
				PipelineCreateRequest: nil,
			},
			wantErr: false,
		},
		{
			name: "test_empty_SnippetConfig",
			fields: &fields{
				PipelineSource:        apistructs.PipelineSourceAutoTest,
				PipelineYmlName:       "test",
				PipelineYml:           "test",
				PipelineCreateRequest: &apistructs.PipelineCreateRequestV2{},
				SnippetConfig:         nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *PipelineDefinitionProcess
			if tt.fields != nil {
				req = &PipelineDefinitionProcess{
					PipelineSource:        tt.fields.PipelineSource,
					PipelineYmlName:       tt.fields.PipelineYmlName,
					PipelineYml:           tt.fields.PipelineYml,
					SnippetConfig:         tt.fields.SnippetConfig,
					PipelineCreateRequest: tt.fields.PipelineCreateRequest,
					VersionLock:           1,
				}
			}

			if err := req.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
