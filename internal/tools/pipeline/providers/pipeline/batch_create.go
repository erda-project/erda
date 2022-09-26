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

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/common/apis"
)

func (s *pipelineService) PipelineBatchCreate(ctx context.Context, req *pb.PipelineBatchCreateRequest) (*pb.PipelineBatchCreateResponse, error) {
	identityInfo := apis.GetIdentityInfo(ctx)
	if req.UserID == "" && identityInfo != nil {
		req.UserID = identityInfo.UserID
	}

	var reqs []*pb.PipelineCreateRequest
	for _, ymlPath := range req.BatchPipelineYmlPaths {
		reqs = append(reqs, &pb.PipelineCreateRequest{
			AppID:           req.AppID,
			Branch:          req.Branch,
			Source:          req.Source,
			PipelineYmlName: ymlPath,
			CallbackURLs:    req.CallbackURLs,
			AutoRun:         req.AutoRun,
			// Creating through the batch interface temporarily only needs to support the gittar type
			PipelineYmlSource: apistructs.PipelineYmlSourceGittar.String(),
			UserID:            req.UserID,
		})
	}

	result := make(map[string]*basepb.PipelineDTO)
	for _, r := range reqs {
		p, err := s.makePipelineFromRequest(r)
		if err != nil {
			return nil, apierrors.ErrBatchCreatePipeline.InternalError(err)
		}
		var stages []spec.PipelineStage
		if stages, err = s.CreatePipelineGraph(p); err != nil {
			return nil, apierrors.ErrBatchCreatePipeline.InternalError(err)
		}

		// PreCheck
		_ = s.PreCheck(p, stages, p.GetUserID(), r.AutoRun)

		// 是否自动执行
		if r.AutoRun {
			if p, err = s.run.RunOnePipeline(ctx, &pb.PipelineRunRequest{
				PipelineID: p.ID,
				UserID:     r.UserID,
				Secrets:    getSecrets(p)},
			); err != nil {
				return nil, apierrors.ErrRunPipeline.InternalError(err)
			}
		}
		result[r.PipelineYmlName] = s.ConvertPipeline(p)
	}

	return &pb.PipelineBatchCreateResponse{
		Data: result,
	}, nil
}

// getSecrets Compatible with bigdata application
func getSecrets(p *spec.Pipeline) map[string]string {
	return map[string]string{
		"gittar.repo":   p.CommitDetail.Repo,
		"gittar.branch": p.Labels[apistructs.LabelBranch],
		"gittar.commit": p.CommitDetail.CommitID,
		"gittar.commit.abbrev": func() string {
			if len(p.CommitDetail.CommitID) > 8 {
				return p.CommitDetail.CommitID[:8]
			}
			return p.CommitDetail.CommitID
		}(),
		"gittar.message": p.CommitDetail.Comment,
		"gittar.author":  p.CommitDetail.Author,
	}
}
