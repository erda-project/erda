package pipelinesvc

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (s *PipelineSvc) RerunFailed(req *apistructs.PipelineRerunFailedRequest) (*spec.Pipeline, error) {
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
		p.Extra.SubmitUser = s.tryGetUser(req.UserID)
	}
	p.Type = apistructs.PipelineTypeRerunFailed

	if err = s.createPipelineGraph(p); err != nil {
		return nil, apierrors.ErrRerunFailedPipeline.InternalError(err)
	}

	// 立即执行一次
	if req.AutoRunAtOnce {
		if p, err = s.RunPipeline(&apistructs.PipelineRunRequest{
			PipelineID:        p.ID,
			IdentityInfo:      req.IdentityInfo,
			PipelineRunParams: origin.Snapshot.RunPipelineParams.ToPipelineRunParams(),
		},
		); err != nil {
			return nil, err
		}
	}

	return p, nil
}
