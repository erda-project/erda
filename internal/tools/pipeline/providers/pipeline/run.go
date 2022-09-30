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
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/common/apis"
)

func (s *pipelineService) PipelineRun(ctx context.Context, req *pb.PipelineRunRequest) (*pb.PipelineRunResponse, error) {
	identityInfo := apis.GetIdentityInfo(ctx)
	if req.UserID == "" && identityInfo != nil {
		req.UserID = identityInfo.UserID
	}
	if req.InternalClient == "" && identityInfo != nil {
		req.InternalClient = identityInfo.InternalClient
	}

	p, err := s.Get(req.PipelineID)
	if err != nil {
		return nil, apierrors.ErrRunPipeline.NotFound()
	}

	if s.edgeRegister.IsCenter() && p.IsEdge {
		s.p.Log.Infof("proxy run pipeline to edge, pipelineID: %d", p.ID)

		if err := s.proxyRunPipelineRequestToEdge(ctx, p, req); err != nil {
			return nil, apierrors.ErrRunPipeline.InternalError(err)
		}

		return &pb.PipelineRunResponse{}, nil
	}

	if _, err := s.run.RunOnePipeline(ctx, req); err != nil {
		return nil, apierrors.ErrRunPipeline.InternalError(err)
	}
	// report
	if s.edgeRegister.IsEdge() {
		s.edgeReporter.TriggerOncePipelineReport(p.ID)
	}
	return &pb.PipelineRunResponse{}, nil
}

func (s *pipelineService) proxyRunPipelineRequestToEdge(ctx context.Context, p *spec.Pipeline, req *pb.PipelineRunRequest) error {
	// handle at edge side
	edgeBundle, err := s.edgeRegister.GetEdgeBundleByClusterName(p.ClusterName)
	if err != nil {
		return err
	}
	return edgeBundle.RunPipeline(*req)
}
