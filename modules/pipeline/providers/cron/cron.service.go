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
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-errors/errors"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	common "github.com/erda-project/erda-proto-go/core/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/cron/db"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

func (s *provider) CronCreate(ctx context.Context, req *pb.CronCreateRequest) (*pb.CronCreateResponse, error) {
	// param validate
	if !apistructs.PipelineSource(req.PipelineSource).Valid() {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("invalid pipelineSource: %s", req.PipelineSource))
	}
	if req.PipelineYmlName == "" {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("missing pipelineYmlName"))
	}
	if req.PipelineYml == "" {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("missing pipelineYml"))
	}

	pipelineYml, err := pipelineyml.New([]byte(req.PipelineYml))
	if err != nil {
		return nil, apierrors.ErrParsePipelineYml.InternalError(err)
	}
	var compensator *apistructs.CronCompensator
	if pipelineYml.Spec().CronCompensator != nil {
		compensator = &apistructs.CronCompensator{}
		compensator.Enable = pipelineYml.Spec().CronCompensator.Enable
		compensator.StopIfLatterExecuted = pipelineYml.Spec().CronCompensator.StopIfLatterExecuted
		compensator.LatestFirst = pipelineYml.Spec().CronCompensator.LatestFirst
	}

	if req.CronExpr == "" {
		req.CronExpr = pipelineYml.Spec().Cron
	}

	createCron := &db.PipelineCron{
		ID:              req.ID,
		PipelineSource:  apistructs.PipelineSource(req.PipelineSource),
		PipelineYmlName: req.PipelineYmlName,
		CronExpr:        req.CronExpr,
		Enable:          &[]bool{req.Enable.Value}[0],
		Extra: db.PipelineCronExtra{
			PipelineYml:            req.PipelineYml,
			ClusterName:            req.ClusterName,
			FilterLabels:           req.FilterLabels,
			NormalLabels:           req.NormalLabels,
			Envs:                   req.Envs,
			ConfigManageNamespaces: req.ConfigManageNamespaces,
			CronStartFrom: func() *time.Time {
				if req.CronStartFrom == nil {
					return nil
				}
				cronStartFrom := req.CronStartFrom.AsTime()
				return &cronStartFrom
			}(),
			Version:         "v2",
			Compensator:     compensator,
			IncomingSecrets: req.IncomingSecrets,
		},
		PipelineDefinitionID: req.PipelineDefinitionID,
	}

	err = Transaction(s.dbClient, func(op mysqlxorm.SessionOption) error {
		if req.CronExpr != "" {
			err = s.InsertOrUpdatePipelineCron(createCron, op)
		} else {
			err = s.disable(createCron, op)
		}
		if err != nil {
			return err
		}
		req.ID = createCron.ID
		if err := s.Daemon.AddIntoPipelineCrond(createCron); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, nil
	}

	return &pb.CronCreateResponse{
		Data: createCron.Convert2DTO(),
	}, nil
}

func Transaction(dbClient *db.Client, do func(option mysqlxorm.SessionOption) error) error {
	txSession := dbClient.NewSession()
	defer txSession.Close()
	if err := txSession.Begin(); err != nil {
		return err
	}
	err := do(mysqlxorm.WithSession(txSession))
	if err != nil {
		if rbErr := txSession.Rollback(); rbErr != nil {
			return err
		}
		return err
	}
	if cmErr := txSession.Commit(); cmErr != nil {
		return cmErr
	}
	return nil
}

func (s *provider) CronPaging(ctx context.Context, req *pb.CronPagingRequest) (*pb.CronPagingResponse, error) {

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

func (s *provider) CronStart(ctx context.Context, req *pb.CronStartRequest) (*pb.CronStartResponse, error) {
	cron, err := s.operate(req.CronID, true)
	if err != nil {
		return nil, err
	}

	return &pb.CronStartResponse{
		Data: cron,
	}, nil
}

func (s *provider) CronStop(ctx context.Context, req *pb.CronStopRequest) (*pb.CronStopResponse, error) {
	cron, err := s.operate(req.CronID, false)
	if err != nil {
		return nil, err
	}

	return &pb.CronStopResponse{
		Data: cron,
	}, nil
}

func (s *provider) operate(cronID uint64, enable bool) (*common.Cron, error) {
	cron, found, err := s.dbClient.GetPipelineCron(cronID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineCron.InternalError(err)
	}
	if !found {
		return nil, apierrors.ErrGetPipelineCron.InternalError(fmt.Errorf("not found"))
	}

	*cron.Enable = enable
	//todo 校验cron.CronExpr是否合法，不合法就不执行
	if enable && cron.CronExpr == "" {
		return cron.Convert2DTO(), nil
	}

	err = Transaction(s.dbClient, func(option mysqlxorm.SessionOption) error {
		if err = s.dbClient.UpdatePipelineCron(cron.ID, &cron, option); err != nil {
			return apierrors.ErrOperatePipeline.InternalError(err)
		}

		if err := s.Daemon.AddIntoPipelineCrond(&cron); err != nil {
			return apierrors.ErrReloadCrond.InternalError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return cron.Convert2DTO(), nil
}

func (s *provider) CronDelete(ctx context.Context, req *pb.CronDeleteRequest) (*pb.CronDeleteResponse, error) {

	err := Transaction(s.dbClient, func(option mysqlxorm.SessionOption) error {
		cron, found, err := s.dbClient.GetPipelineCron(req.CronID, option)
		if err != nil {
			return apierrors.ErrDeletePipelineCron.InvalidParameter(err)
		}
		if !found {
			return apierrors.ErrDeletePipelineCron.InvalidParameter(fmt.Errorf("not found"))
		}

		if err := s.dbClient.DeletePipelineCron(cron.ID, option); err != nil {
			return apierrors.ErrDeletePipelineCron.InternalError(err)
		}

		if err := s.Daemon.DeleteFromPipelineCrond(&cron); err != nil {
			return apierrors.ErrDeletePipelineCron.InternalError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &pb.CronDeleteResponse{}, nil
}

func (s *provider) CronGet(ctx context.Context, req *pb.CronGetRequest) (*pb.CronGetResponse, error) {
	cron, found, err := s.dbClient.GetPipelineCron(req.CronID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineCron.InvalidParameter(err)
	}
	if !found {
		return &pb.CronGetResponse{
			Data: nil,
		}, nil
	}

	return &pb.CronGetResponse{
		Data: cron.Convert2DTO(),
	}, nil
}

func (s *provider) CronUpdate(ctx context.Context, req *pb.CronUpdateRequest) (*pb.CronUpdateResponse, error) {
	cron, found, err := s.dbClient.GetPipelineCron(req.CronID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("not found")
	}

	pipeline, err := pipelineyml.New([]byte(req.PipelineYml))
	if err != nil {
		return nil, err
	}
	if pipeline.Spec().CronCompensator != nil {
		cron.Extra.Compensator = &apistructs.CronCompensator{
			Enable:               pipeline.Spec().CronCompensator.Enable,
			LatestFirst:          pipeline.Spec().CronCompensator.LatestFirst,
			StopIfLatterExecuted: pipeline.Spec().CronCompensator.StopIfLatterExecuted,
		}
	}
	cron.CronExpr = req.CronExpr
	cron.Extra.PipelineYml = req.PipelineYml
	cron.Extra.ConfigManageNamespaces = strutil.DedupSlice(append(cron.Extra.ConfigManageNamespaces, req.ConfigManageNamespaces...), true)
	cron.Extra.IncomingSecrets = req.Secrets
	var fields = []string{spec.PipelineCronCronExpr, spec.Extra}
	if req.PipelineDefinitionID != "" {
		cron.PipelineDefinitionID = req.PipelineDefinitionID
		fields = append(fields, spec.PipelineDefinitionID)
	}

	err = Transaction(s.dbClient, func(option mysqlxorm.SessionOption) error {
		err = s.dbClient.UpdatePipelineCronWillUseDefault(cron.ID, &cron, fields, option)
		if err != nil {
			return apierrors.ErrUpdatePipelineCron.InternalError(err)
		}

		if err := s.Daemon.AddIntoPipelineCrond(&cron); err != nil {
			return apierrors.ErrUpdatePipelineCron.InternalError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &pb.CronUpdateResponse{}, nil
}

func (s *provider) InsertOrUpdatePipelineCron(new *db.PipelineCron, ops ...mysqlxorm.SessionOption) error {
	var err error

	// 寻找 v1
	queryV1 := &db.PipelineCron{
		ApplicationID:   new.ApplicationID,
		Branch:          new.Branch,
		PipelineYmlName: new.PipelineYmlName,
	}
	v1Exist, err := s.dbClient.GetDBClient().Get(queryV1)
	if err != nil {
		return err
	}
	if queryV1.Extra.Version == "v2" {
		v1Exist = false
	}
	if v1Exist {
		new.ID = queryV1.ID
		new.Enable = queryV1.Enable
		err := s.dbClient.UpdatePipelineCron(new.ID, new, ops...)
		if err != nil {
			return err
		}
		return nil
	}

	// 寻找 v2
	queryV2 := &db.PipelineCron{
		PipelineSource:  new.PipelineSource,
		PipelineYmlName: new.PipelineYmlName,
	}
	v2Exist, err := s.dbClient.GetDBClient().Get(queryV2)
	if err != nil {
		return err
	}
	if v2Exist {
		new.ID = queryV2.ID
		new.Enable = queryV2.Enable
		err := s.dbClient.UpdatePipelineCron(new.ID, new, ops...)
		if err != nil {
			return err
		}
		return nil
	}

	err = s.dbClient.CreatePipelineCron(new, ops...)
	if err != nil {
		return err
	}
	return nil
}

func (s *provider) disable(cron *db.PipelineCron, option mysqlxorm.SessionOption) error {
	var disable = false
	var updateCron = &db.PipelineCron{}
	var columns = []string{spec.PipelineCronCronExpr, spec.PipelineCronEnable}
	var err error

	queryV1 := &db.PipelineCron{
		ApplicationID:   cron.ApplicationID,
		Branch:          cron.Branch,
		PipelineYmlName: cron.PipelineYmlName,
	}
	v1Exist, err := s.dbClient.GetDBClient().Get(queryV1)
	if err != nil {
		return err
	}
	if queryV1.Extra.Version == "v2" {
		v1Exist = false
	}
	if v1Exist {
		updateCron.Enable = &disable
		updateCron.ID = queryV1.ID
		updateCron.CronExpr = cron.CronExpr
		err := s.dbClient.UpdatePipelineCronWillUseDefault(updateCron.ID, updateCron, columns, option)
		if err != nil {
			return err
		}
		cron.ID = updateCron.ID
		return nil
	}

	queryV2 := &db.PipelineCron{
		PipelineSource:  cron.PipelineSource,
		PipelineYmlName: cron.PipelineYmlName,
	}
	v2Exist, err := s.dbClient.GetDBClient().Get(queryV2)
	if err != nil {
		return err
	}
	if v2Exist {
		updateCron.Enable = &disable
		updateCron.ID = queryV2.ID
		updateCron.CronExpr = cron.CronExpr
		err := s.dbClient.UpdatePipelineCronWillUseDefault(updateCron.ID, updateCron, columns, option)
		if err != nil {
			return err
		}
		cron.ID = updateCron.ID
		return nil
	}
	return nil
}

func pbCronToDBCron(pbCron *common.Cron) (*db.PipelineCron, error) {
	dbCronJson, err := json.Marshal(pbCron)
	if err != nil {
		return nil, err
	}

	var result db.PipelineCron
	err = json.Unmarshal(dbCronJson, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
