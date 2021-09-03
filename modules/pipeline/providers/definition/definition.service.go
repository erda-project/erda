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
	context "context"

	pb "github.com/erda-project/erda-proto-go/core/pipeline/definition/pb"
	"github.com/erda-project/erda/modules/pipeline/providers/definition/db"
	"github.com/erda-project/erda/modules/pipeline/providers/definition/transform_type"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

type definitionService struct {
	p        *provider
	dbClient *db.Client
}

func (s *definitionService) Process(ctx context.Context, req *pb.PipelineDefinitionProcessRequest) (*pb.PipelineDefinitionProcessResponse, error) {
	if !apis.IsInternalClient(ctx) {
		return nil, apierrors.ErrrProcessPipelineDefinition.AccessDenied()
	}
	var bo = &transform_type.PipelineDefinitionProcess{}
	if err := bo.ReqTransform(req); err != nil {
		return nil, err
	}

	result, err := s.ProcessPipelineDefinition(ctx, bo)
	if err != nil {
		return nil, err
	}

	return result.TransformToResp()
}

func (s *definitionService) Version(ctx context.Context, req *pb.PipelineDefinitionProcessVersionRequest) (*pb.PipelineDefinitionProcessVersionResponse, error) {
	if !apis.IsInternalClient(ctx) {
		return nil, apierrors.ErrrProcessPipelineDefinition.AccessDenied()
	}

	var bo = &transform_type.GetPipelineDefinitionVersion{}
	bo.ReqTransform(req)

	return s.GetPipelineDefinitionVersionLock(bo)
}
