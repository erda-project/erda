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

package pipelinecronsvc

import (
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/services/crondsvc"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

type PipelineCronSvc struct {
	dbClient *dbclient.Client
	crondSvc *crondsvc.CrondSvc
}

func New(dbClient *dbclient.Client, crondSvc *crondsvc.CrondSvc) *PipelineCronSvc {
	s := PipelineCronSvc{}
	s.dbClient = dbClient
	s.crondSvc = crondSvc
	return &s
}

func (s *PipelineCronSvc) Paging(req apistructs.PipelineCronPagingRequest) ([]spec.PipelineCron, int64, error) {
	return s.dbClient.PagingPipelineCron(req)
}

func (s *PipelineCronSvc) Start(cronID uint64) (*spec.PipelineCron, error) {
	return s.operate(cronID, true)
}

func (s *PipelineCronSvc) Stop(cronID uint64) (*spec.PipelineCron, error) {
	return s.operate(cronID, false)
}

func (s *PipelineCronSvc) operate(cronID uint64, enable bool) (*spec.PipelineCron, error) {
	cron, err := s.dbClient.GetPipelineCron(cronID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineCron.InternalError(err)
	}

	*cron.Enable = enable
	//todo 校验cron.CronExpr是否合法，不合法就不执行
	if enable && cron.CronExpr == "" {
		return &cron, nil
	}
	if err = s.dbClient.UpdatePipelineCron(cron.ID, &cron); err != nil {
		return nil, apierrors.ErrOperatePipeline.InternalError(err)
	}

	if err := s.crondSvc.AddIntoPipelineCrond(cron.ID); err != nil {
		return nil, apierrors.ErrReloadCrond.InternalError(err)
	}

	return &cron, nil
}

func (s *PipelineCronSvc) Create(req apistructs.PipelineCronCreateRequest) (*spec.PipelineCron, error) {
	// param validate
	if !req.PipelineCreateRequest.PipelineSource.Valid() {
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
	cron := spec.PipelineCron{
		PipelineSource:  req.PipelineCreateRequest.PipelineSource,
		PipelineYmlName: req.PipelineCreateRequest.PipelineYmlName,
		CronExpr:        pipelineYml.Spec().Cron,
		Enable:          &[]bool{req.PipelineCreateRequest.AutoStartCron}[0],
		Extra: spec.PipelineCronExtra{
			PipelineYml:   req.PipelineCreateRequest.PipelineYml,
			ClusterName:   req.PipelineCreateRequest.ClusterName,
			FilterLabels:  req.PipelineCreateRequest.Labels,
			NormalLabels:  req.PipelineCreateRequest.NormalLabels,
			Envs:          req.PipelineCreateRequest.Envs,
			CronStartFrom: req.PipelineCreateRequest.CronStartFrom,
			Version:       "v2",
		},
	}
	err = s.dbClient.InsertOrUpdatePipelineCron(&cron)
	if err != nil {
		return nil, apierrors.ErrCreatePipelineCron.InternalError(err)
	}

	if err := s.crondSvc.AddIntoPipelineCrond(cron.ID); err != nil {
		return nil, apierrors.ErrReloadCrond.InternalError(err)
	}

	return &cron, nil
}

func (s *PipelineCronSvc) Delete(cronID uint64) error {
	cron, err := s.dbClient.GetPipelineCron(cronID)
	if err != nil {
		return apierrors.ErrDeletePipelineCron.InvalidParameter(err)
	}
	if err := s.dbClient.DeletePipelineCron(cron.ID); err != nil {
		return apierrors.ErrDeletePipelineCron.InternalError(err)
	}
	if err := s.crondSvc.DeletePipelineCrond(cron.ID); err != nil {
		return apierrors.ErrReloadCrond.InternalError(err)
	}
	return nil
}

func (s *PipelineCronSvc) Get(cronID uint64) (*spec.PipelineCron, error) {
	cron, err := s.dbClient.GetPipelineCron(cronID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineCron.InvalidParameter(err)
	}
	return &cron, nil
}

// PipelineCronUpdate pipeline cron update
func (s *PipelineCronSvc) PipelineCronUpdate(req apistructs.PipelineCronUpdateRequest) error {
	cron, err := s.dbClient.GetPipelineCron(req.ID)
	if err != nil {
		return err
	}
	cron.CronExpr = req.CronExpr
	cron.Extra.PipelineYml = req.PipelineYml
	err = s.dbClient.UpdatePipelineCronWillUseDefault(cron.ID, &cron, []string{spec.PipelineCronCronExpr, spec.Extra})
	return err
}
