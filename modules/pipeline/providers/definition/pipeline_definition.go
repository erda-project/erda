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

package definition_client

import (
	"context"

	"github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda/modules/pipeline/providers/definition/db"
)

type pipelineDefinition struct {
	dbClient *db.Client
}

func (p pipelineDefinition) Create(ctx context.Context, request *pb.PipelineDefinitionCreateRequest) (*pb.PipelineDefinitionCreateResponse, error) {
	panic("implement me")
}

func (p pipelineDefinition) Update(ctx context.Context, request *pb.PipelineDefinitionUpdateRequest) (*pb.PipelineDefinitionUpdateResponse, error) {
	panic("implement me")
}

func (p pipelineDefinition) Delete(ctx context.Context, request *pb.PipelineDefinitionDeleteRequest) (*pb.PipelineDefinitionDeleteResponse, error) {
	panic("implement me")
}

func (p pipelineDefinition) Get(ctx context.Context, request *pb.PipelineDefinitionGetRequest) (*pb.PipelineDefinitionGetResponse, error) {
	panic("implement me")
}

func (p pipelineDefinition) List(ctx context.Context, request *pb.PipelineDefinitionListRequest) (*pb.PipelineDefinitionListResponse, error) {
	panic("implement me")
}


