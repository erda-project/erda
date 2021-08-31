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

package definition_adaptor

import (
	"context"
	"reflect"
	"testing"

	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/definition/transform_type"
)

type server struct {
}

func (s server) Process(ctx context.Context, request *pb.PipelineDefinitionProcessRequest) (*pb.PipelineDefinitionProcessResponse, error) {
	return &pb.PipelineDefinitionProcessResponse{
		Id:              1,
		PipelineSource:  request.PipelineSource,
		PipelineYmlName: request.PipelineYmlName,
		PipelineYml:     "version: 1.1",
		VersionLock:     1,
	}, nil
}

func (s server) Version(ctx context.Context, request *pb.PipelineDefinitionProcessVersionRequest) (*pb.PipelineDefinitionProcessVersionResponse, error) {
	return &pb.PipelineDefinitionProcessVersionResponse{
		VersionLock: 1,
	}, nil
}

func TestProvider_ProcessPipelineDefinition(t *testing.T) {
	type args struct {
		req transform_type.ClientPipelineDefinitionProcessRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *transform_type.ClientPipelineDefinitionResponse
		wantErr bool
	}{
		{
			name: "test_value",
			args: args{
				req: transform_type.ClientPipelineDefinitionProcessRequest{
					PipelineSource:  apistructs.PipelineSourceDice,
					PipelineYml:     "version: 1.1",
					PipelineYmlName: "test",
					SnippetConfig: &apistructs.SnippetConfig{
						Name:   "test",
						Source: "test",
					},
					PipelineCreateRequest: &apistructs.PipelineCreateRequestV2{
						PipelineYmlName: "test",
					},
					VersionLock: 1,
				},
			},
			want: &transform_type.ClientPipelineDefinitionResponse{
				ID:              1,
				PipelineYmlName: "test",
				PipelineSource:  "dice",
				PipelineYml:     "version: 1.1",
				VersionLock:     1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{
				ClientDefinitionService: server{},
			}
			got, err := p.ProcessPipelineDefinition(context.Background(), tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessPipelineDefinition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProcessPipelineDefinition() got = %v, want %v", got, tt.want)
			}
		})
	}
}
