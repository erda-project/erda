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

package cron

import (
	context "context"

	"github.com/go-errors/errors"

	pb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	common "github.com/erda-project/erda-proto-go/core/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/cron/db"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/services/crondsvc"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

type Service struct {
	p        *provider
	dbClient *db.Client
	crondSvc *crondsvc.CrondSvc
}

func (p *Service) WithPipelineSvc(crondSvc *crondsvc.CrondSvc) {
	p.crondSvc = crondSvc
}

func (s *Service) CronCreate(ctx context.Context, req *pb.CronCreateRequest) (*pb.CronCreateResponse, error) {
	// param validate
	if !apistructs.PipelineSource(req.PipelineCreateRequest.PipelineSource).Valid() {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("invalid pipelineSource: %s", req.PipelineCreateRequest.PipelineSource))
	}
	if req.PipelineCreateRequest.PipelineYmlName == "" {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("missing pipelineYmlName"))
	}
	if req.PipelineCreateRequest.PipelineYml == "" {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("missing pipelineYml"))
	}
	pipelineYml, err := pipelineyml.New([]byte(req.PipelineCreateRequest.PipelineYml))
	if err != nil {
		return nil, err
	}
	if pipelineYml.Spec().Cron == "" {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("not cron pipeline"))
	}

	// store to db
	cron := db.PipelineCron{
		PipelineSource:  apistructs.PipelineSource(req.PipelineCreateRequest.PipelineSource),
		PipelineYmlName: req.PipelineCreateRequest.PipelineYmlName,
		CronExpr:        pipelineYml.Spec().Cron,
		Enable:          &[]bool{req.PipelineCreateRequest.AutoStartCron}[0],
	}
	if req.PipelineCreateRequest != nil {
		var extra = db.PipelineCronExtra{
			PipelineYml:  req.PipelineCreateRequest.PipelineYml,
			ClusterName:  req.PipelineCreateRequest.ClusterName,
			FilterLabels: req.PipelineCreateRequest.Labels,
			NormalLabels: req.PipelineCreateRequest.NormalLabels,
			Envs:         req.PipelineCreateRequest.Envs,
			Version:      "v2",
		}
		if req.PipelineCreateRequest.CronStartFrom != nil {
			var cronStartFrom = req.PipelineCreateRequest.CronStartFrom.AsTime()
			extra.CronStartFrom = &cronStartFrom
		}
		cron.Extra = extra
	}

	err = s.dbClient.InsertOrUpdatePipelineCron(&cron)
	if err != nil {
		return nil, apierrors.ErrCreatePipelineCron.InternalError(err)
	}

	if err := s.crondSvc.AddIntoPipelineCrond(cron.ID); err != nil {
		return nil, apierrors.ErrReloadCrond.InternalError(err)
	}

	return &pb.CronCreateResponse{
		Data: cron.ID,
	}, nil
}

func (s *Service) CronPaging(ctx context.Context, req *pb.CronPagingRequest) (*pb.CronPagingResponse, error) {

	crons, total, err := s.dbClient.PagingPipelineCron(req)
	if err != nil {
		return nil, err
	}

	var data []*common.Cron
	for _, c := range crons {
		data = append(data, c.Convert2DTO())
	}

	result := pb.CronPagingResponse{
		Total: total,
		Data:  data,
	}

	return &result, nil
}

func (s *Service) CronStart(ctx context.Context, req *pb.CronStartRequest) (*pb.CronStartResponse, error) {
	cron, err := s.operate(req.CronID, true)
	if err != nil {
		return nil, err
	}

	return &pb.CronStartResponse{
		Data: cron,
	}, nil
}

func (s *Service) CronStop(ctx context.Context, req *pb.CronStopRequest) (*pb.CronStopResponse, error) {
	cron, err := s.operate(req.CronID, false)
	if err != nil {
		return nil, err
	}

	return &pb.CronStopResponse{
		Data: cron,
	}, nil
}

func (s *Service) operate(cronID uint64, enable bool) (*common.Cron, error) {
	cron, err := s.dbClient.GetPipelineCron(cronID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineCron.InternalError(err)
	}

	*cron.Enable = enable
	//todo 校验cron.CronExpr是否合法，不合法就不执行
	if enable && cron.CronExpr == "" {
		return cron.Convert2DTO(), nil
	}

	if err = s.dbClient.UpdatePipelineCron(cron.ID, &cron); err != nil {
		return nil, apierrors.ErrOperatePipeline.InternalError(err)
	}

	if err := s.crondSvc.AddIntoPipelineCrond(cron.ID); err != nil {
		return nil, apierrors.ErrReloadCrond.InternalError(err)
	}

	return cron.Convert2DTO(), nil
}

func (s *Service) CronDelete(ctx context.Context, req *pb.CronDeleteRequest) (*pb.CronDeleteResponse, error) {

	cron, err := s.dbClient.GetPipelineCron(req.CronID)
	if err != nil {
		return nil, apierrors.ErrDeletePipelineCron.InvalidParameter(err)
	}

	if err := s.dbClient.DeletePipelineCron(cron.ID); err != nil {
		return nil, apierrors.ErrDeletePipelineCron.InternalError(err)
	}

	if err := s.crondSvc.DeletePipelineCrond(cron.ID); err != nil {
		return nil, apierrors.ErrReloadCrond.InternalError(err)
	}

	return &pb.CronDeleteResponse{}, nil
}

func (s *Service) CronGet(ctx context.Context, req *pb.CronGetRequest) (*pb.CronGetResponse, error) {
	cron, err := s.dbClient.GetPipelineCron(req.CronID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineCron.InvalidParameter(err)
	}

	return &pb.CronGetResponse{
		Data: cron.Convert2DTO(),
	}, nil
}

func (s *Service) CronUpdate(ctx context.Context, req *pb.CronUpdateRequest) (*pb.CronUpdateResponse, error) {
	cron, err := s.dbClient.GetPipelineCron(req.CronID)
	if err != nil {
		return nil, err
	}

	cron.CronExpr = req.CronExpr
	cron.Extra.PipelineYml = req.PipelineYml
	cron.Extra.ConfigManageNamespaces = strutil.DedupSlice(append(cron.Extra.ConfigManageNamespaces, req.ConfigManageNamespaces...), true)

	var fields = []string{spec.PipelineCronCronExpr, spec.Extra}

	if req.PipelineDefinitionID != "" {
		cron.PipelineDefinitionID = req.PipelineDefinitionID
		fields = append(fields, spec.PipelineDefinitionID)
	}
	err = s.dbClient.UpdatePipelineCronWillUseDefault(cron.ID, &cron, fields)
	if err != nil {
		return nil, err
	}

	return &pb.CronUpdateResponse{}, nil
}
