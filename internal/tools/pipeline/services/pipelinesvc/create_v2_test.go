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

package pipelinesvc

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"

	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	dpb "github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pb"
	sourcepb "github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/pkg/mock"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/edgepipeline_register"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/edgereporter"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type CronServiceServerTestImpl struct {
	create *cronpb.CronCreateResponse
}

func (c CronServiceServerTestImpl) CronCreate(ctx context.Context, request *cronpb.CronCreateRequest) (*cronpb.CronCreateResponse, error) {
	return c.create, nil
}

func (c CronServiceServerTestImpl) CronPaging(ctx context.Context, request *cronpb.CronPagingRequest) (*cronpb.CronPagingResponse, error) {
	panic("implement me")
}

func (c CronServiceServerTestImpl) CronStart(ctx context.Context, request *cronpb.CronStartRequest) (*cronpb.CronStartResponse, error) {
	panic("implement me")
}

func (c CronServiceServerTestImpl) CronStop(ctx context.Context, request *cronpb.CronStopRequest) (*cronpb.CronStopResponse, error) {
	panic("implement me")
}

func (c CronServiceServerTestImpl) CronDelete(ctx context.Context, request *cronpb.CronDeleteRequest) (*cronpb.CronDeleteResponse, error) {
	panic("implement me")
}

func (c CronServiceServerTestImpl) CronGet(ctx context.Context, request *cronpb.CronGetRequest) (*cronpb.CronGetResponse, error) {
	panic("implement me")
}

func (c CronServiceServerTestImpl) CronUpdate(ctx context.Context, request *cronpb.CronUpdateRequest) (*cronpb.CronUpdateResponse, error) {
	panic("implement me")
}

func TestPipelineSvc_UpdatePipelineCron(t *testing.T) {
	type args struct {
		p                      *spec.Pipeline
		cronStartFrom          *time.Time
		configManageNamespaces []string
		cronCompensator        *pipelineyml.CronCompensator
	}
	tests := []struct {
		name    string
		args    args
		isEdge  bool
		wantErr bool
	}{
		{
			name: "test id > 0",
			args: args{
				p: &spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						PipelineSource:  "test",
						PipelineYmlName: "test",
						ClusterName:     "test",
					},
					PipelineExtra: spec.PipelineExtra{
						Extra: spec.PipelineExtraInfo{
							CronExpr: "test",
						},
						PipelineYml: "test",
						Snapshot: spec.Snapshot{
							Envs: map[string]string{
								"test": "test",
							},
						},
					},
					Labels: map[string]string{
						"test": "test",
					},
				},
				cronStartFrom:          nil,
				configManageNamespaces: nil,
				cronCompensator:        nil,
			},
			isEdge:  false,
			wantErr: false,
		},
		{
			name: "test edge cron",
			args: args{
				p: &spec.Pipeline{
					PipelineBase: spec.PipelineBase{
						PipelineSource:  "test",
						PipelineYmlName: "test",
						ClusterName:     "test",
					},
					PipelineExtra: spec.PipelineExtra{
						Extra: spec.PipelineExtraInfo{
							CronExpr: "test",
						},
						PipelineYml: "test",
						Snapshot: spec.Snapshot{
							Envs: map[string]string{
								"test": "test",
							},
						},
					},
					Labels: map[string]string{
						"test": "test",
					},
				},
				cronStartFrom:          nil,
				configManageNamespaces: nil,
				cronCompensator:        nil,
			},
			isEdge:  true,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PipelineSvc{}
			impl := CronServiceServerTestImpl{
				create: &cronpb.CronCreateResponse{
					Data: &pb.Cron{
						ID: 1,
					},
				},
			}
			s.pipelineCronSvc = impl
			if tt.isEdge {
				s.edgeRegister = &edgepipeline_register.MockEdgeRegister{}
				s.edgeReporter = &edgereporter.MockEdgeReporter{}
			}

			if err := s.UpdatePipelineCron(tt.args.p, tt.args.cronStartFrom, tt.args.configManageNamespaces, tt.args.cronCompensator); (err != nil) != tt.wantErr {
				t.Errorf("UpdatePipelineCron() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parseSourceDicePipelineYmlName(t *testing.T) {
	type args struct {
		ymlName string
		branch  string
	}
	tests := []struct {
		name string
		args args
		want *pipelineYmlName
	}{
		{
			name: "test with other source",
			args: args{
				ymlName: "pipeline.yml",
				branch:  "master",
			},
			want: nil,
		},
		{
			name: "test with dice source1",
			args: args{
				ymlName: "200/DEV/develop/pipeline.yml",
				branch:  "develop",
			},
			want: &pipelineYmlName{
				appID:     "200",
				workspace: "DEV",
				branch:    "develop",
				fileName:  "pipeline.yml",
			},
		},
		{
			name: "test with dice source2",
			args: args{
				ymlName: "200/DEV/develop/.erda/pipeline.yml",
				branch:  "develop",
			},
			want: &pipelineYmlName{
				appID:     "200",
				workspace: "DEV",
				branch:    "develop",
				fileName:  ".erda/pipeline.yml",
			},
		},
		{
			name: "test with dice source3",
			args: args{
				ymlName: "200/DEV/feature/erda/.erda/pipelines/pipeline.yml",
				branch:  "feature/erda",
			},
			want: &pipelineYmlName{
				appID:     "200",
				workspace: "DEV",
				branch:    "feature/erda",
				fileName:  ".erda/pipelines/pipeline.yml",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseSourceDicePipelineYmlName(tt.args.ymlName, tt.args.branch); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSourceDicePipelineYmlName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getFilePath(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with getFilePath1",
			args: args{
				path: "pipeline.yml",
			},
			want: "",
		},
		{
			name: "test with getFilePath2",
			args: args{
				path: ".erda/pipelines/pipeline.yml",
			},
			want: ".erda/pipelines",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFilePath(tt.args.path); got != tt.want {
				t.Errorf("getFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_makePipelineName(t *testing.T) {
	type args struct {
		pipelineYml string
		fileName    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test with name from pipelineYml",
			args: args{
				pipelineYml: `
version: "1.1"
name: pipeline-deploy
stages:
`,
				fileName: "pipeline.yml",
			},
			want: "pipeline-deploy",
		},
		{
			name: "test with name from fileName",
			args: args{
				pipelineYml: `
version: "1.1"
stages:
`,
				fileName: "pipeline.yml",
			},
			want: "pipeline.yml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makePipelineName(tt.args.pipelineYml, tt.args.fileName); got != tt.want {
				t.Errorf("makePipelineName() = %v, want %v", got, tt.want)
			}
		})
	}
}

type sourceMock struct {
	mock.SourceMock
}

func (s sourceMock) Create(ctx context.Context, request *sourcepb.PipelineSourceCreateRequest) (*sourcepb.PipelineSourceCreateResponse, error) {
	if request.Ref == "master" {
		return nil, fmt.Errorf("error")
	}
	return &sourcepb.PipelineSourceCreateResponse{
		PipelineSource: &sourcepb.PipelineSource{ID: "0008b4dd-95c0-4b56-ace7-8b358d8d5895"},
	}, nil
}

type definitionMock struct {
	mock.DefinitionMock
}

func (d definitionMock) Create(ctx context.Context, request *dpb.PipelineDefinitionCreateRequest) (*dpb.PipelineDefinitionCreateResponse, error) {
	return &dpb.PipelineDefinitionCreateResponse{PipelineDefinition: &dpb.PipelineDefinition{
		ID:               "679033d3-f047-43de-880e-52bda6c7a792",
		PipelineSourceID: "0008b4dd-95c0-4b56-ace7-8b358d8d5895",
	}}, nil
}

func TestPipelineSvc_GetDefinitionID(t *testing.T) {
	var bdl *bundle.Bundle
	monkey.PatchInstanceMethod(reflect.TypeOf(bdl), "GetApp", func(*bundle.Bundle, uint64) (*apistructs.ApplicationDTO, error) {
		return &apistructs.ApplicationDTO{
			ID:          1,
			Name:        "erda",
			OrgName:     "terminus",
			ProjectName: "reda-project",
		}, nil
	})
	defer monkey.UnpatchAll()

	type fields struct {
		pipelineSource     sourcepb.SourceServiceServer
		pipelineDefinition dpb.DefinitionServiceServer
		bdl                *bundle.Bundle
	}
	type args struct {
		req *apistructs.PipelineCreateRequestV2
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test with from req DefinitionID",
			fields: fields{
				pipelineSource:     sourceMock{},
				pipelineDefinition: definitionMock{},
				bdl:                bdl,
			},
			args: args{
				req: &apistructs.PipelineCreateRequestV2{
					DefinitionID: "aceb2c20-d76d-4d0a-bbea-6161b9a931ba",
				},
			},
			want:    "aceb2c20-d76d-4d0a-bbea-6161b9a931ba",
			wantErr: false,
		},
		{
			name: "test with other source",
			fields: fields{
				pipelineSource:     sourceMock{},
				pipelineDefinition: definitionMock{},
				bdl:                bdl,
			},
			args: args{
				req: &apistructs.PipelineCreateRequestV2{
					PipelineSource: "auto-test",
				},
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "test with create with dice",
			fields: fields{
				pipelineSource:     sourceMock{},
				pipelineDefinition: definitionMock{},
				bdl:                bdl,
			},
			args: args{
				req: &apistructs.PipelineCreateRequestV2{
					PipelineYmlName: "1/DEV/develop/pipeline.yml",
					PipelineSource:  "dice",
					Labels:          map[string]string{"branch": "develop"},
				},
			},
			want:    "679033d3-f047-43de-880e-52bda6c7a792",
			wantErr: false,
		},
		{
			name: "test with create with dice with error",
			fields: fields{
				pipelineSource:     sourceMock{},
				pipelineDefinition: definitionMock{},
				bdl:                bdl,
			},
			args: args{
				req: &apistructs.PipelineCreateRequestV2{
					PipelineYmlName: "1/DEV/master/pipeline.yml",
					PipelineSource:  "dice",
					Labels:          map[string]string{"branch": "master"},
				},
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &PipelineSvc{
				pipelineSource:     tt.fields.pipelineSource,
				pipelineDefinition: tt.fields.pipelineDefinition,
				bdl:                tt.fields.bdl,
			}
			got, err := s.GetDefinitionID(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDefinitionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDefinitionID() got = %v, want %v", got, tt.want)
			}
		})
	}
}
