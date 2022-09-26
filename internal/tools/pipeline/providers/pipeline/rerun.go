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

	"google.golang.org/protobuf/types/known/timestamppb"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	common "github.com/erda-project/erda-proto-go/core/pipeline/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/cron/crontypes"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/common/apis"
)

func (s *pipelineService) PipelineRerun(ctx context.Context, req *pb.PipelineRerunRequest) (*pb.PipelineRerunResponse, error) {
	identityInfo := apis.GetIdentityInfo(ctx)
	if req.UserID == "" && identityInfo != nil {
		req.UserID = identityInfo.UserID
	}
	if req.InternalClient == "" && identityInfo != nil {
		req.InternalClient = identityInfo.InternalClient
	}

	p, err := s.Get(req.PipelineID)
	if err != nil {
		return nil, apierrors.ErrRerunPipeline.NotFound()
	}

	canProxy := s.edgeRegister.CanProxyToEdge(p.PipelineSource, p.ClusterName)
	if canProxy {
		s.p.Log.Infof("proxy rerun pipeline to edge, pipelineID: %d", p.ID)
		pipelineDto, err := s.proxyRerunPipelineRequestToEdge(ctx, p, req)
		if err != nil {
			return nil, apierrors.ErrRunPipeline.InternalError(err)
		}
		return &pb.PipelineRerunResponse{Data: pipelineDto}, nil
	}

	newP, err := s.Rerun(ctx, req)
	if s.edgeRegister.IsEdge() {
		s.edgeReporter.TriggerOncePipelineReport(newP.ID)
	}

	return &pb.PipelineRerunResponse{Data: s.ConvertPipeline(newP)}, nil
}

// Rerun commit unchanged
func (s *pipelineService) Rerun(ctx context.Context, req *pb.PipelineRerunRequest) (*spec.Pipeline, error) {

	origin, err := s.dbClient.GetPipeline(req.PipelineID)
	if err != nil {
		return nil, apierrors.ErrRerunPipeline.InternalError(err)
	}

	var originCron *common.Cron
	if origin.CronID != nil {
		result, err := s.cronSvc.CronGet(ctx, &cronpb.CronGetRequest{
			CronID: *origin.CronID,
		})
		if err != nil {
			return nil, apierrors.ErrRerunPipeline.InternalError(err)
		}
		if result.Data == nil {
			return nil, apierrors.ErrNotFoundPipelineCron.InternalError(crontypes.ErrCronNotFound)
		}
		originCron = result.Data
	}

	if origin.Labels == nil {
		origin.Labels = map[string]string{}
	}
	origin.Labels[apistructs.LabelPipelineType] = apistructs.PipelineTypeRerun.String()

	runParams, err := origin.Snapshot.RunPipelineParams.ToPipelineRunParamsPB()
	if err != nil {
		return nil, err
	}
	p, err := s.CreateV2(ctx, &pb.PipelineCreateRequestV2{
		PipelineYml:            origin.PipelineYml,
		ClusterName:            origin.ClusterName,
		PipelineYmlName:        origin.PipelineYmlName,
		RunParams:              runParams,
		PipelineSource:         origin.PipelineSource.String(),
		Labels:                 origin.Labels,
		NormalLabels:           origin.GenerateNormalLabelsForCreateV2(),
		Envs:                   origin.Snapshot.Envs,
		ConfigManageNamespaces: origin.GetConfigManageNamespaces(),
		AutoRunAtOnce:          req.AutoRunAtOnce,
		AutoStartCron:          false,
		CronStartFrom: func() *timestamppb.Timestamp {
			if originCron == nil {
				return nil
			}
			cronStartFrom := originCron.Extra.CronStartFrom.AsTime()
			return timestamppb.New(cronStartFrom)
		}(),
		UserID:         req.UserID,
		InternalClient: req.InternalClient,
		DefinitionID:   origin.PipelineDefinitionID,
		Secrets:        req.Secrets,
	})
	if err != nil {
		return nil, err
	}

	p.Extra.CopyFromPipelineID = &origin.ID

	if err := s.dbClient.UpdatePipelineExtraExtraInfoByPipelineID(p.ID, p.Extra); err != nil {
		return nil, apierrors.ErrUpdatePipeline.InternalError(err)
	}

	return p, nil
}

func (s *pipelineService) proxyRerunPipelineRequestToEdge(ctx context.Context, p *spec.Pipeline, req *pb.PipelineRerunRequest) (*basepb.PipelineDTO, error) {
	// handle at edge side
	edgeBundle, err := s.edgeRegister.GetEdgeBundleByClusterName(p.ClusterName)
	if err != nil {
		return nil, err
	}
	return edgeBundle.RerunPipeline(*req)
}
