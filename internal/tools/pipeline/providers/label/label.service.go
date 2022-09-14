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

package label

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-proto-go/core/pipeline/label/pb"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/common/apis"
)

type labelService struct {
	p *provider

	impl *labelServiceImpl
}

func (s *labelService) PipelineLabelBatchInsert(ctx context.Context, req *pb.PipelineLabelBatchInsertRequest) (*pb.PipelineLabelBatchInsertResponse, error) {
	if !apis.IsInternalClient(ctx) {
		return nil, fmt.Errorf("auth error: not internal client")
	}
	if len(req.Labels) == 0 {
		return nil, apierrors.ErrCreatePipelineLabel.InvalidParameter("labels")
	}
	for index, label := range req.Labels {
		if label.TargetID <= 0 {
			return nil, apierrors.ErrCreatePipelineLabel.InvalidParameter(fmt.Errorf("label index %v missing targetID", index))
		}
		if len(label.PipelineYmlName) <= 0 {
			return nil, apierrors.ErrCreatePipelineLabel.InvalidParameter(fmt.Errorf("label index %v missing pipelineYmlName", index))
		}
		if len(label.PipelineSource) <= 0 {
			return nil, apierrors.ErrCreatePipelineLabel.InvalidParameter(fmt.Errorf("label index %v missing pipelineSource", index))
		}
		if len(label.Type) <= 0 {
			return nil, apierrors.ErrCreatePipelineLabel.InvalidParameter(fmt.Errorf("label index %v missing type", index))
		}
		if len(label.Key) <= 0 {
			return nil, apierrors.ErrCreatePipelineLabel.InvalidParameter(fmt.Errorf("label index %v missing key", index))
		}
	}
	err := s.impl.BatchCreateLabels(req)
	if err != nil {
		return nil, apierrors.ErrCreatePipelineLabel.InternalError(err)
	}
	return &pb.PipelineLabelBatchInsertResponse{}, nil
}

func (s *labelService) PipelineLabelList(ctx context.Context, req *pb.PipelineLabelListRequest) (*pb.PipelineLabelListResponse, error) {
	if !apis.IsInternalClient(ctx) {
		return nil, fmt.Errorf("auth error: not internal client")
	}

	if len(req.PipelineYmlName) <= 0 {
		return nil, apierrors.ErrListPipelineLabel.InvalidParameter("missing pipelineYmlName")
	}

	if len(req.PipelineSource) <= 0 {
		return nil, apierrors.ErrListPipelineLabel.InvalidParameter("missing pipelineSource")
	}

	pageResult, err := s.impl.ListLabels(req)
	if err != nil {
		return nil, err
	}

	return &pb.PipelineLabelListResponse{
		Data: pageResult,
	}, nil
}
