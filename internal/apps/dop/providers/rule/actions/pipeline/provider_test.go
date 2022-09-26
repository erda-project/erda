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

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/services/branchrule"
	"github.com/erda-project/erda/internal/apps/dop/services/pipeline"
)

func Test_provider_CreatePipeline(t *testing.T) {
	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetApp",
		func(d *bundle.Bundle, id uint64) (*apistructs.ApplicationDTO, error) {
			return &apistructs.ApplicationDTO{ID: 1}, nil
		},
	)

	defer monkey.UnpatchAll()

	p := &provider{
		bdl: bdl,
	}
	monkey.PatchInstanceMethod(reflect.TypeOf(p), "GetPipelineDefinitionID",
		func(p *provider, ctx context.Context, app *apistructs.ApplicationDTO, branch, path, name, strPipelineYml string) (definitionID string, err error) {
			return "1", nil
		},
	)

	var pipelineSvc *pipeline.Pipeline
	monkey.PatchInstanceMethod(reflect.TypeOf(pipelineSvc), "ConvertPipelineToV2",
		func(p *pipeline.Pipeline, pv1 *pipelinepb.PipelineCreateRequest) (*pipelinepb.PipelineCreateRequestV2, error) {
			return &pipelinepb.PipelineCreateRequestV2{}, nil
		},
	)

	monkey.PatchInstanceMethod(reflect.TypeOf(pipelineSvc), "CreatePipelineV2",
		func(p *pipeline.Pipeline, reqPipeline *pipelinepb.PipelineCreateRequestV2) (*basepb.PipelineDTO, error) {
			return &basepb.PipelineDTO{ID: 1}, nil
		},
	)
	var branch *branchrule.BranchRule
	monkey.PatchInstanceMethod(reflect.TypeOf(branch), "Query",
		func(p *branchrule.BranchRule, scopeType apistructs.ScopeType, scopeID int64) ([]*apistructs.BranchRule, error) {
			return []*apistructs.BranchRule{{ID: 1}}, nil
		},
	)

	p.pipeline = pipelineSvc
	p.branchRule = branch
	type args struct {
		env map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			args: args{
				env: map[string]interface{}{
					"git_push": map[string]interface{}{
						"pipelineConfig": &PipelineConfig{
							RefName: "feature/123",
						},
					},
				},
			},
			want: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := p.CreatePipeline(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("provider.CreatePipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("provider.CreatePipeline() = %v, want %v", got, tt.want)
			}
		})
	}
}
