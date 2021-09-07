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

package pipeline

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/definition_client/deftype"
)

func TestGetBranch(t *testing.T) {
	ss := []struct {
		ref  string
		Want string
	}{
		{"", ""},
		{"refs/heads/", ""},
		{"refs/heads/master", "master"},
		{"refs/heads/feature/test", "feature/test"},
	}
	for _, v := range ss {
		assert.Equal(t, v.Want, getBranch(v.ref))
	}
}

func TestIsPipelineYmlPath(t *testing.T) {
	ss := []struct {
		path string
		want bool
	}{
		{"pipeline.yml", true},
		{".dice/pipelines/a.yml", true},
		{"", false},
		{"dice/pipeline.yml", false},
	}
	for _, v := range ss {
		assert.Equal(t, v.want, isPipelineYmlPath(v.path))
	}

}

type process struct{}

func (process) ProcessPipelineDefinition(ctx context.Context, req deftype.ClientDefinitionProcessRequest) (*deftype.ClientDefinitionProcessResponse, error) {
	return &deftype.ClientDefinitionProcessResponse{
		ID:              1,
		PipelineSource:  req.PipelineSource,
		PipelineYmlName: req.PipelineYmlName,
		PipelineYml:     req.PipelineYml,
		VersionLock:     req.VersionLock,
	}, nil
}

func TestPipeline_deletePipelineDefinition(t *testing.T) {
	type args struct {
		name   string
		appID  uint64
		branch string
		userID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				name:   "test",
				appID:  1,
				branch: "test",
				userID: "1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipeline{
				ds: process{},
			}
			patch := monkey.PatchInstanceMethod(reflect.TypeOf(p), "ConvertPipelineToV2", func(p *Pipeline, pv1 *apistructs.PipelineCreateRequest) (*apistructs.PipelineCreateRequestV2, error) {
				return &apistructs.PipelineCreateRequestV2{
					PipelineYmlName: pv1.PipelineYmlName,
					PipelineSource:  apistructs.PipelineSource(pv1.PipelineYmlSource),
					PipelineYml:     pv1.PipelineYmlContent,
				}, nil
			})
			if err := p.deletePipelineDefinition(tt.args.name, tt.args.appID, tt.args.branch, tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("deletePipelineDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
			patch.Unpatch()
		})
	}
}

func TestPipeline_reportPipelineDefinition(t *testing.T) {
	type args struct {
		appDto      *apistructs.ApplicationDTO
		userID      string
		branch      string
		name        string
		pipelineYml string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				appDto: &apistructs.ApplicationDTO{
					Name:        "test",
					ProjectID:   1,
					ProjectName: "test",
					OrgID:       1,
				},
				branch:      "test",
				userID:      "1",
				name:        "test",
				pipelineYml: "version: 1.1",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Pipeline{
				ds: process{},
			}
			patch := monkey.PatchInstanceMethod(reflect.TypeOf(p), "ConvertPipelineToV2", func(p *Pipeline, pv1 *apistructs.PipelineCreateRequest) (*apistructs.PipelineCreateRequestV2, error) {
				return &apistructs.PipelineCreateRequestV2{
					PipelineYmlName: pv1.PipelineYmlName,
					PipelineSource:  apistructs.PipelineSource(pv1.PipelineYmlSource),
					PipelineYml:     pv1.PipelineYmlContent,
				}, nil
			})

			if err := p.reportPipelineDefinition(tt.args.appDto, tt.args.userID, tt.args.branch, tt.args.name, tt.args.pipelineYml); (err != nil) != tt.wantErr {
				t.Errorf("reportPipelineDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
			patch.Unpatch()
		})
	}
}
