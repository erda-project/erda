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

package mock

import (
	"context"

	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
)

type DefinitionMock struct {
}

func (d DefinitionMock) Create(ctx context.Context, request *pb.PipelineDefinitionCreateRequest) (*pb.PipelineDefinitionCreateResponse, error) {
	panic("implement me")
}

func (d DefinitionMock) Update(ctx context.Context, request *pb.PipelineDefinitionUpdateRequest) (*pb.PipelineDefinitionUpdateResponse, error) {
	panic("implement me")
}

func (d DefinitionMock) Delete(ctx context.Context, request *pb.PipelineDefinitionDeleteRequest) (*pb.PipelineDefinitionDeleteResponse, error) {
	panic("implement me")
}

func (d DefinitionMock) DeleteByRemote(ctx context.Context, request *pb.PipelineDefinitionDeleteByRemoteRequest) (*pb.PipelineDefinitionDeleteResponse, error) {
	panic("implement me")
}

func (d DefinitionMock) Get(ctx context.Context, request *pb.PipelineDefinitionGetRequest) (*pb.PipelineDefinitionGetResponse, error) {
	panic("implement me")
}

func (d DefinitionMock) List(ctx context.Context, request *pb.PipelineDefinitionListRequest) (*pb.PipelineDefinitionListResponse, error) {
	panic("implement me")
}

func (d DefinitionMock) StatisticsGroupByRemote(ctx context.Context, request *pb.PipelineDefinitionStatisticsRequest) (*pb.PipelineDefinitionStatisticsResponse, error) {
	panic("implement me")
}

func (d DefinitionMock) ListUsedRefs(ctx context.Context, request *pb.PipelineDefinitionUsedRefListRequest) (*pb.PipelineDefinitionUsedRefListResponse, error) {
	panic("implement me")
}

func (d DefinitionMock) StatisticsGroupByFilePath(ctx context.Context, request *pb.PipelineDefinitionStatisticsRequest) (*pb.PipelineDefinitionStatisticsResponse, error) {
	panic("implement me")
}

func (d DefinitionMock) UpdateExtra(ctx context.Context, request *pb.PipelineDefinitionExtraUpdateRequest) (*pb.PipelineDefinitionExtraUpdateResponse, error) {
	panic("implement me")
}

func (d DefinitionMock) ListByRemote(ctx context.Context, request *pb.PipelineDefinitionListByRemoteRequest) (*pb.PipelineDefinitionListResponse, error) {
	panic("implement me")
}
