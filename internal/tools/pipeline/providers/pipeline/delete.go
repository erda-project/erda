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

	commonpb "github.com/erda-project/erda-proto-go/common/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

func (s *pipelineService) PipelineDelete(ctx context.Context, req *pb.PipelineDeleteRequest) (*pb.PipelineDeleteResponse, error) {
	// get details for authentication
	p, err := s.Get(req.PipelineID)
	if err != nil {
		return nil, apierrors.ErrDeletePipeline.NotFound()
	}

	// check whether the user has the OPERATE permission under the corresponding branch of the application
	identityInfo := apis.GetIdentityInfo(ctx)
	if err := s.permission.CheckBranch(&commonpb.IdentityInfo{
		UserID:         identityInfo.UserID,
		InternalClient: identityInfo.InternalClient,
	}, p.Labels[apistructs.LabelAppID], p.Labels[apistructs.LabelBranch], apistructs.OperateAction); err != nil {
		return nil, apierrors.ErrDeletePipeline.AccessDenied()
	}

	if err := s.Delete(req.PipelineID); err != nil {
		return nil, apierrors.ErrDeletePipeline.InternalError(err)
	}

	return &pb.PipelineDeleteResponse{}, nil
}

func (s *pipelineService) Delete(pipelineID uint64) error {

	// 获取 pipeline
	p, err := s.Get(pipelineID)
	if err != nil {
		return apierrors.ErrGetPipeline.InvalidParameter(err)
	}
	// 校验当前流水线是否可被删除
	can, reason := canDelete(*p)
	if !can {
		return apierrors.ErrDeletePipeline.InvalidState(reason)
	}

	// pipelines
	if err := s.dbClient.DeletePipeline(pipelineID); err != nil {
		return apierrors.ErrDeletePipeline.InternalError(err)
	}

	// related pipeline stages
	if err := s.dbClient.DeletePipelineStagesByPipelineID(pipelineID); err != nil {
		return apierrors.ErrDeletePipelineStage.InternalError(err)
	}

	// related pipeline tasks
	if err := s.dbClient.DeletePipelineTasksByPipelineID(pipelineID); err != nil {
		return apierrors.ErrDeletePipelineTask.InternalError(err)
	}

	// related pipeline labels
	if err := s.dbClient.DeletePipelineLabelsByPipelineID(pipelineID); err != nil {
		return apierrors.ErrDeletePipelineLabel.InternalError(err)
	}

	return nil
}
