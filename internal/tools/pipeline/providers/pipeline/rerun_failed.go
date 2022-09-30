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
	"fmt"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/pkg/errors"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/common/apis"
)

func (s *pipelineService) PipelineRerunFailed(ctx context.Context, req *pb.PipelineRerunFailedRequest) (*pb.PipelineRerunFailedResponse, error) {
	identityInfo := apis.GetIdentityInfo(ctx)
	if req.UserID == "" && identityInfo != nil {
		req.UserID = identityInfo.UserID
	}
	if req.InternalClient == "" && identityInfo != nil {
		req.InternalClient = identityInfo.InternalClient
	}

	p, err := s.Get(req.PipelineID)
	if err != nil {
		return nil, apierrors.ErrRerunFailedPipeline.NotFound()
	}

	if s.edgeRegister.IsCenter() && p.IsEdge {
		s.p.Log.Infof("proxy rerun-failed pipeline to edge, pipelineID: %d", p.ID)
		pipelineDto, err := s.proxyRerunFailedPipelineRequestToEdge(ctx, p, req)
		if err != nil {
			return nil, apierrors.ErrRerunFailedPipeline.InternalError(err)
		}
		return &pb.PipelineRerunFailedResponse{Data: pipelineDto}, nil
	}

	newP, err := s.RerunFailed(ctx, req)
	if err != nil {
		return nil, err
	}

	// report
	if s.edgeRegister.IsEdge() {
		s.edgeReporter.TriggerOncePipelineReport(p.ID)
	}

	return &pb.PipelineRerunFailedResponse{Data: s.ConvertPipeline(newP)}, nil
}

func (s *pipelineService) RerunFailed(ctx context.Context, req *pb.PipelineRerunFailedRequest) (*spec.Pipeline, error) {
	// base pipeline
	origin, err := s.dbClient.GetPipeline(req.PipelineID)
	if err != nil {
		return nil, apierrors.ErrRerunFailedPipeline.InternalError(err)
	}

	if origin.Extra.CompleteReconcilerGC {
		return nil, apierrors.ErrRerunFailedPipeline.InvalidState("cannot rerun, already complete gc")
	}

	if !origin.Status.IsEndStatus() {
		return nil, apierrors.ErrRerunFailedPipeline.InvalidState("cannot rerun, not end status")
	}

	// 寻找上一次的失败节点
	rerunFailedDetail, err := s.dbClient.FindCauseFailedPipelineTasks(origin.ID)
	if err != nil {
		return nil, apierrors.ErrRerunFailedPipeline.InternalError(err)
	}

	p, err := s.makePipelineFromCopy(&origin)
	if err != nil {
		return nil, apierrors.ErrRerunFailedPipeline.InternalError(err)
	}

	// 重试失败节点必须在同一个集群
	if origin.ClusterName != p.ClusterName {
		return nil, apierrors.ErrRerunFailedPipeline.InvalidState(fmt.Sprintf(
			"cannot rerun pipeline in another cluster, before: %s, now: %s", origin.ClusterName, p.ClusterName))
	}

	p.Extra.RerunFailedDetail = &rerunFailedDetail
	if req.UserID != "" {
		p.Extra.SubmitUser = s.user.TryGetUser(ctx, req.UserID)
	}
	p.Type = apistructs.PipelineTypeRerunFailed

	var stages []spec.PipelineStage
	if stages, err = s.CreatePipelineGraph(p); err != nil {
		return nil, apierrors.ErrRerunFailedPipeline.InternalError(err)
	}
	// PreCheck
	_ = s.PreCheck(p, stages, p.GetUserID(), req.AutoRunAtOnce)

	// 立即执行一次
	if req.AutoRunAtOnce {
		runParams, err := origin.Snapshot.RunPipelineParams.ToPipelineRunParamsPB()
		if err != nil {
			return nil, err
		}
		if p, err = s.run.RunOnePipeline(ctx, &pb.PipelineRunRequest{
			PipelineID:        p.ID,
			UserID:            req.UserID,
			InternalClient:    req.InternalClient,
			PipelineRunParams: runParams,
			Secrets:           req.Secrets,
		},
		); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (s *pipelineService) makePipelineFromCopy(o *spec.Pipeline) (p *spec.Pipeline, err error) {
	r := deepcopy.Copy(o)
	p, ok := r.(*spec.Pipeline)
	if !ok {
		return nil, errors.New("failed to copy pipeline")
	}

	now := time.Now()

	// 初始化一些字段
	p.ID = 0
	p.Status = apistructs.PipelineStatusAnalyzed
	p.PipelineExtra.PipelineID = 0
	p.Snapshot = spec.Snapshot{}
	p.Snapshot.Envs = o.Snapshot.Envs
	p.Snapshot.RunPipelineParams = o.Snapshot.RunPipelineParams
	p.Extra.Namespace = o.Extra.Namespace
	p.Extra.SubmitUser = &basepb.PipelineUser{}
	p.Extra.RunUser = &basepb.PipelineUser{}
	p.Extra.CancelUser = &basepb.PipelineUser{}
	p.Extra.ShowMessage = nil
	p.Extra.CopyFromPipelineID = &o.ID
	p.Extra.RerunFailedDetail = nil
	p.Extra.CronTriggerTime = nil
	p.Extra.CompleteReconcilerGC = false
	p.TriggerMode = apistructs.PipelineTriggerModeManual // 手动触发
	p.TimeCreated = &now
	p.TimeUpdated = &now
	p.TimeBegin = nil
	p.TimeEnd = nil
	p.CostTimeSec = -1
	p.PipelineDefinitionID = o.PipelineDefinitionID

	return p, nil
}

func (s *pipelineService) proxyRerunFailedPipelineRequestToEdge(ctx context.Context, p *spec.Pipeline, req *pb.PipelineRerunFailedRequest) (*basepb.PipelineDTO, error) {
	// handle at edge side
	edgeBundle, err := s.edgeRegister.GetEdgeBundleByClusterName(p.ClusterName)
	if err != nil {
		return nil, err
	}
	return edgeBundle.RerunFailedPipeline(*req)
}
