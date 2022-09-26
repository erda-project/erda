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

	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

type OperateAction string

var (
	OpDisableTask OperateAction = "DISABLE-TASK"
	OpEnableTask  OperateAction = "ENABLE-TASK"
	OpPauseTask   OperateAction = "PAUSE-TASK"
	OpUnpauseTask OperateAction = "UNPAUSE-TASK"
)

func (s *pipelineService) PipelineOperate(ctx context.Context, req *pb.PipelineOperateRequest) (*pb.PipelineOperateResponse, error) {
	identityInfo := apis.GetIdentityInfo(ctx)
	if err := s.checkPipelineOperatePermission(identityInfo, req.PipelineID, req, apistructs.OperateAction); err != nil {
		return nil, apierrors.ErrOperatePipeline.AccessDenied()
	}

	if err := s.Operate(req.PipelineID, req); err != nil {
		return nil, apierrors.ErrOperatePipeline.InternalError(err)
	}

	return &pb.PipelineOperateResponse{}, nil
}

func (s *pipelineService) Operate(pipelineID uint64, req *pb.PipelineOperateRequest) error {
	if req == nil || len(req.TaskOperates) <= 0 {
		return nil
	}

	p, err := s.dbClient.GetPipeline(pipelineID)
	if err != nil {
		return apierrors.ErrOperatePipeline.InternalError(err)
	}

	extra := p.PipelineExtra
	extra.Extra.TaskOperates = append(extra.Extra.TaskOperates, req.TaskOperates...)

	return s.dbClient.UpdatePipelineExtraExtraInfoByPipelineID(p.ID, extra.Extra)
}
