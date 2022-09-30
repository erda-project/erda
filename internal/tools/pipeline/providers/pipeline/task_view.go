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
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
)

func (s *pipelineService) PipelineTaskView(ctx context.Context, req *pb.PipelineTaskViewRequest) (*pb.PipelineTaskViewResponse, error) {
	if req.PipelineID != 0 {
		detail, err := s.Detail(req.PipelineID)
		if err != nil {
			return nil, apierrors.ErrTaskView.InternalError(err)
		}
		return &pb.PipelineTaskViewResponse{Data: detail}, nil
	}
	if len(req.YmlNames) == 0 {
		return nil, apierrors.ErrTaskView.MissingParameter("ymlNames")
	}
	if len(req.Sources) == 0 {
		return nil, apierrors.ErrTaskView.MissingParameter("sources")
	}

	var condition pb.PipelinePagingRequest
	condition.YmlName = []string{req.YmlNames}
	condition.Source = []string{req.Sources}
	condition.PageNum = 1
	condition.PageSize = 1
	pageResult, err := s.List(&condition)
	if err != nil {
		return nil, apierrors.ErrTaskView.InternalError(err)
	}
	if len(pageResult.Pipelines) == 0 {
		return nil, apierrors.ErrTaskView.NotFound()
	}

	detail, err := s.Detail(pageResult.Pipelines[0].ID)
	if err != nil {
		return nil, apierrors.ErrTaskView.InternalError(err)
	}
	return &pb.PipelineTaskViewResponse{Data: detail}, nil
}
