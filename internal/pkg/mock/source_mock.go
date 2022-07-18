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

	"github.com/erda-project/erda-proto-go/core/pipeline/source/pb"
)

type SourceMock struct {
}

func (s SourceMock) Create(ctx context.Context, request *pb.PipelineSourceCreateRequest) (*pb.PipelineSourceCreateResponse, error) {
	panic("implement me")
}

func (s SourceMock) Update(ctx context.Context, request *pb.PipelineSourceUpdateRequest) (*pb.PipelineSourceUpdateResponse, error) {
	panic("implement me")
}

func (s SourceMock) Delete(ctx context.Context, request *pb.PipelineSourceDeleteRequest) (*pb.PipelineSourceDeleteResponse, error) {
	panic("implement me")
}

func (s SourceMock) Get(ctx context.Context, request *pb.PipelineSourceGetRequest) (*pb.PipelineSourceGetResponse, error) {
	panic("implement me")
}

func (s SourceMock) List(ctx context.Context, request *pb.PipelineSourceListRequest) (*pb.PipelineSourceListResponse, error) {
	panic("implement me")
}

func (s SourceMock) DeleteByRemote(ctx context.Context, request *pb.PipelineSourceDeleteByRemoteRequest) (*pb.PipelineSourceDeleteResponse, error) {
	panic("implement me")
}
