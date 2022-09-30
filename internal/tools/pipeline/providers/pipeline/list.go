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
)

func (s *pipelineService) PipelinePaging(ctx context.Context, req *pb.PipelinePagingRequest) (*pb.PipelinePagingResponse, error) {
	err := apistructs.PostHandlePBQueryString(req)
	if err != nil {
		return nil, apierrors.ErrListPipeline.InvalidParameter(err)
	}
	pageResult, err := s.List(req)
	if err != nil {
		return nil, apierrors.ErrListPipeline.InternalError(err)
	}
	return &pb.PipelinePagingResponse{
		Data: pageResult,
	}, nil
}

func (s *pipelineService) List(req *pb.PipelinePagingRequest) (*pb.PipelineListResponseData, error) {
	pagingResult, err := s.dbClient.PageListPipelines(req)
	if err != nil {
		return nil, err
	}
	pipelines := pagingResult.Pipelines
	total := pagingResult.Total
	currentPageSize := pagingResult.CurrentPageSize
	var result pb.PipelineListResponseData
	result.Pipelines = s.BatchConvert2PagePipeline(pipelines)
	result.Total = total
	result.CurrentPageSize = currentPageSize

	return &result, nil
}
