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

package endpoints

import (
	"context"
	"reflect"
	"testing"

	"bou.ke/monkey"
	"github.com/pkg/errors"

	defpb "github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	sourcepd "github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
)

func Test_checkApplicationCreateParam(t *testing.T) {
	tests := []struct {
		name string
		req  apistructs.ApplicationCreateRequest
		want error
	}{
		{
			name: "invalid_name",
			req:  apistructs.ApplicationCreateRequest{},
			want: errors.Errorf("invalid request, name is empty"),
		},
		{
			name: "invalid_projectID",
			req: apistructs.ApplicationCreateRequest{
				Name: "demo",
			},
			want: errors.Errorf("invalid request, projectId is empty"),
		},
		{
			name: "invalid_mode",
			req: apistructs.ApplicationCreateRequest{
				Name:      "demo",
				ProjectID: uint64(1),
				Mode:      "app",
			},
			want: errors.New("invalid mode"),
		},
	}
	for _, tt := range tests {
		if err := checkApplicationCreateParam(tt.req); err.Error() != tt.want.Error() {
			t.Errorf("checkApplicationCreateParam() = %v, want %v", err, tt.want)
		}
	}
}

type MockPipelineSource struct{}

func (m MockPipelineSource) Create(ctx context.Context, request *sourcepd.PipelineSourceCreateRequest) (*sourcepd.PipelineSourceCreateResponse, error) {
	panic("implement me")
}

func (m MockPipelineSource) Update(ctx context.Context, request *sourcepd.PipelineSourceUpdateRequest) (*sourcepd.PipelineSourceUpdateResponse, error) {
	panic("implement me")
}

func (m MockPipelineSource) Delete(ctx context.Context, request *sourcepd.PipelineSourceDeleteRequest) (*sourcepd.PipelineSourceDeleteResponse, error) {
	panic("implement me")
}

func (m MockPipelineSource) Get(ctx context.Context, request *sourcepd.PipelineSourceGetRequest) (*sourcepd.PipelineSourceGetResponse, error) {
	panic("implement me")
}

func (m MockPipelineSource) List(ctx context.Context, request *sourcepd.PipelineSourceListRequest) (*sourcepd.PipelineSourceListResponse, error) {
	panic("implement me")
}

func (m MockPipelineSource) DeleteByRemote(ctx context.Context, request *sourcepd.PipelineSourceDeleteByRemoteRequest) (*sourcepd.PipelineSourceDeleteResponse, error) {
	return nil, nil
}

type MockPipelineDefinition struct{}

func (m MockPipelineDefinition) Create(ctx context.Context, request *defpb.PipelineDefinitionCreateRequest) (*defpb.PipelineDefinitionCreateResponse, error) {
	panic("implement me")
}

func (m MockPipelineDefinition) Update(ctx context.Context, request *defpb.PipelineDefinitionUpdateRequest) (*defpb.PipelineDefinitionUpdateResponse, error) {
	panic("implement me")
}

func (m MockPipelineDefinition) Delete(ctx context.Context, request *defpb.PipelineDefinitionDeleteRequest) (*defpb.PipelineDefinitionDeleteResponse, error) {
	panic("implement me")
}

func (m MockPipelineDefinition) DeleteByRemote(ctx context.Context, request *defpb.PipelineDefinitionDeleteByRemoteRequest) (*defpb.PipelineDefinitionDeleteResponse, error) {
	return nil, nil
}

func (m MockPipelineDefinition) Get(ctx context.Context, request *defpb.PipelineDefinitionGetRequest) (*defpb.PipelineDefinitionGetResponse, error) {
	panic("implement me")
}

func (m MockPipelineDefinition) List(ctx context.Context, request *defpb.PipelineDefinitionListRequest) (*defpb.PipelineDefinitionListResponse, error) {
	panic("implement me")
}

func (m MockPipelineDefinition) StatisticsGroupByRemote(ctx context.Context, request *defpb.PipelineDefinitionStatisticsRequest) (*defpb.PipelineDefinitionStatisticsResponse, error) {
	panic("implement me")
}

func (m MockPipelineDefinition) ListUsedRefs(ctx context.Context, request *defpb.PipelineDefinitionUsedRefListRequest) (*defpb.PipelineDefinitionUsedRefListResponse, error) {
	panic("implement me")
}

func (m MockPipelineDefinition) StatisticsGroupByFilePath(ctx context.Context, request *defpb.PipelineDefinitionStatisticsRequest) (*defpb.PipelineDefinitionStatisticsResponse, error) {
	panic("implement me")
}

func (m MockPipelineDefinition) UpdateExtra(ctx context.Context, request *defpb.PipelineDefinitionExtraUpdateRequest) (*defpb.PipelineDefinitionExtraUpdateResponse, error) {
	panic("implement me")
}

func TestEndpoints_deletePipelineSourceAndDefinition(t *testing.T) {
	type fields struct {
		bdl                *bundle.Bundle
		PipelineSource     sourcepd.SourceServiceServer
		PipelineDefinition defpb.DefinitionServiceServer
	}
	type args struct {
		ctx     context.Context
		project *apistructs.ProjectDTO
		app     *apistructs.ApplicationDTO
	}
	source := MockPipelineSource{}
	definition := MockPipelineDefinition{}

	bdl := &bundle.Bundle{}
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetOrg", func(*bundle.Bundle, interface{}) (*apistructs.OrgDTO, error) {
		return &apistructs.OrgDTO{
			ID:   1,
			Name: "erda",
		}, nil
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test with error",
			fields: fields{
				bdl:                bdl,
				PipelineSource:     source,
				PipelineDefinition: definition,
			},
			args: args{
				ctx:     context.Background(),
				project: nil,
				app:     nil,
			},
			wantErr: true,
		},
		{
			name: "test with delete app",
			fields: fields{
				bdl:                bdl,
				PipelineSource:     source,
				PipelineDefinition: definition,
			},
			args: args{
				ctx:     context.Background(),
				project: nil,
				app: &apistructs.ApplicationDTO{
					ID:          1,
					Name:        "app",
					OrgName:     "project",
					ProjectName: "erda",
				},
			},
			wantErr: false,
		},
		{
			name: "test with delete project",
			fields: fields{
				bdl:                bdl,
				PipelineSource:     source,
				PipelineDefinition: definition,
			},
			args: args{
				ctx: context.Background(),
				project: &apistructs.ProjectDTO{
					ID:   1,
					Name: "project",
				},
				app: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Endpoints{
				bdl:                tt.fields.bdl,
				PipelineSource:     tt.fields.PipelineSource,
				PipelineDefinition: tt.fields.PipelineDefinition,
			}
			if err := e.deletePipelineSourceAndDefinition(tt.args.ctx, tt.args.project, tt.args.app); (err != nil) != tt.wantErr {
				t.Errorf("deletePipelineSourceAndDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
