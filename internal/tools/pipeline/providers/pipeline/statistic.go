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

func (s *pipelineService) PipelineStatistic(ctx context.Context, req *pb.PipelineStatisticRequest) (*pb.PipelineStatisticResponse, error) {
	data, err := s.Statistic(req.Source, req.ClusterName)
	if err != nil {
		return nil, apierrors.ErrStatisticPipeline.InternalError(err)
	}
	return &pb.PipelineStatisticResponse{Data: data}, nil
}

// Statistic pipeline operation statistics
func (s *pipelineService) Statistic(source, clusterName string) (*pb.PipelineStatisticResponseData, error) {
	return s.dbClient.PipelineStatistic(source, clusterName)
}
