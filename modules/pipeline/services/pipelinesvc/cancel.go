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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
)

func (s *PipelineSvc) Cancel(ctx context.Context, req *apistructs.PipelineCancelRequest) error {

	p, err := s.dbClient.GetPipeline(req.PipelineID)
	if err != nil {
		return apierrors.ErrGetPipeline.InternalError(err)
	}

	// support idempotent cancel
	if p.Status.IsEndStatus() {
		return nil
	}

	// pipeline 状态判断
	if !p.IsSnippet && !p.Status.CanCancel() {
		return errors.Errorf("invalid status [%s]", p.Status)
	}

	// 设置 cancel user
	if req.UserID != "" {
		p.Extra.CancelUser = s.tryGetUser(req.UserID)
		if err := s.dbClient.UpdatePipelineExtraExtraInfoByPipelineID(p.ID, p.Extra); err != nil {
			return err
		}
	}

	return s.engine.DistributedStopPipeline(ctx, p.ID)
}
