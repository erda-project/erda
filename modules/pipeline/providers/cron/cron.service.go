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
	"fmt"
	"time"

	"github.com/go-errors/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	pb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	common "github.com/erda-project/erda-proto-go/core/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/providers/cron/db"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

func (s *provider) CronCreate(ctx context.Context, req *pb.CronCreateRequest) (*pb.CronCreateResponse, error) {
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
	cron := &db.PipelineCron{
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
			Compensator: &apistructs.CronCompensator{
				Enable: pipelineYml.Spec().CronCompensator.Enable,
			},
		}
		extra.ConfigManageNamespaces = strutil.DedupSlice(append(cron.Extra.ConfigManageNamespaces, req.PipelineCreateRequest.ConfigManageNamespaces...), true)
		extra.IncomingSecrets = req.PipelineCreateRequest.GetSecrets()

		if pipelineYml.Spec().CronCompensator != nil {
			extra.Compensator = &apistructs.CronCompensator{
				Enable:               pipelineYml.Spec().CronCompensator.Enable,
				LatestFirst:          pipelineYml.Spec().CronCompensator.LatestFirst,
				StopIfLatterExecuted: pipelineYml.Spec().CronCompensator.StopIfLatterExecuted,
			}
		}

		if req.PipelineCreateRequest.CronStartFrom != nil {
			var cronStartFrom = req.PipelineCreateRequest.CronStartFrom.AsTime()
			extra.CronStartFrom = &cronStartFrom
		}
		cron.Extra = extra
	}

	// tx
	txSession := s.dbClient.NewSession()
	defer txSession.Close()
	if err := txSession.Begin(); err != nil {
		return nil, err
	}

	err = func() error {
		err = s.InsertOrUpdatePipelineCron(cron, mysqlxorm.WithSession(txSession))
		if err != nil {
			return apierrors.ErrCreatePipelineCron.InternalError(err)
		}

		if err := s.Daemon.AddIntoPipelineCrond(cron); err != nil {
			return apierrors.ErrReloadCrond.InternalError(err)
		}
		return nil
	}()
	if err != nil {
		if rbErr := txSession.Rollback(); rbErr != nil {
			return nil, err
		}
		return nil, err
	}
	if cmErr := txSession.Commit(); cmErr != nil {
		return nil, cmErr
	}

	return &pb.CronCreateResponse{
		Data: cron.ID,
	}, nil
}

func (s *provider) InsertOrUpdatePipelineCron(new_ *db.PipelineCron, ops ...mysqlxorm.SessionOption) error {
	var bdl *bundle.Bundle
	var err error

	toEdge := s.EdgePipelineRegister.ShouldDispatchToEdge(new_.PipelineSource.String(), new_.Extra.ClusterName)
	if toEdge {
		new_.IsEdge = true
		bdl, err = s.EdgePipelineRegister.GetEdgeBundleByClusterName(new_.Extra.ClusterName)
		if err != nil {
			return fmt.Errorf("failed to GetEdgeBundleByClusterName error %v", err)
		}
	}

	// 寻找 v1
	queryV1 := &db.PipelineCron{
		ApplicationID:   new_.ApplicationID,
		Branch:          new_.Branch,
		PipelineYmlName: new_.PipelineYmlName,
	}
	v1Exist, err := s.dbClient.DB().Get(queryV1)
	if err != nil {
		return err
	}
	if queryV1.Extra.Version == "v2" {
		v1Exist = false
	}
	if v1Exist {
		new_.ID = queryV1.ID
		new_.Enable = queryV1.Enable
		err := s.dbClient.UpdatePipelineCron(new_.ID, new_, ops...)
		if err != nil {
			return err
		}

		if toEdge {
			err := bdl.EdgeCronUpdate(&pb.EdgeCronUpdateRequest{
				Cron: new_.Convert2DTO(),
			})
			if err != nil {
				return err
			}
		}
		return nil
	}

	// 寻找 v2
	queryV2 := &db.PipelineCron{
		PipelineSource:  new_.PipelineSource,
		PipelineYmlName: new_.PipelineYmlName,
	}
	v2Exist, err := s.dbClient.DB().Get(queryV2)
	if err != nil {
		return err
	}
	if v2Exist {
		new_.ID = queryV2.ID
		new_.Enable = queryV2.Enable
		err := s.dbClient.UpdatePipelineCron(new_.ID, new_, ops...)
		if err != nil {
			return err
		}

		if toEdge {
			err := bdl.EdgeCronUpdate(&pb.EdgeCronUpdateRequest{
				Cron: new_.Convert2DTO(),
			})
			if err != nil {
				return err
			}
		}
		return nil
	}

	err = s.dbClient.CreatePipelineCron(new_, ops...)
	if err != nil {
		return err
	}
	if toEdge {
		_, err := bdl.EdgeCronCreate(&pb.EdgeCronCreateRequest{
			Cron: new_.Convert2DTO(),
		})
		if err != nil {
			return err
		}
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

	// tx
	txSession := s.dbClient.NewSession()
	defer txSession.Close()
	if err := txSession.Begin(); err != nil {
		return nil, err
	}

	err = func() error {
		toEdge := s.EdgePipelineRegister.ShouldDispatchToEdge(cron.PipelineSource.String(), cron.Extra.ClusterName)
		if toEdge {
			cron.IsEdge = true
		}

		if err = s.dbClient.UpdatePipelineCron(cron.ID, &cron, mysqlxorm.WithSession(txSession)); err != nil {
			return apierrors.ErrOperatePipeline.InternalError(err)
		}

		if err := s.Daemon.AddIntoPipelineCrond(&cron); err != nil {
			return apierrors.ErrReloadCrond.InternalError(err)
		}

		if toEdge {
			bdl, err := s.EdgePipelineRegister.GetEdgeBundleByClusterName(cron.Extra.ClusterName)
			if err != nil {
				return fmt.Errorf("failed to GetEdgeBundleByClusterName error %v", err)
			}

			if enable {
				_, err = bdl.CronStart(cron.ID)
			} else {
				_, err = bdl.CronStop(cron.ID)
			}
			if err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		if rbErr := txSession.Rollback(); rbErr != nil {
			return nil, err
		}
		return nil, err
	}
	if cmErr := txSession.Commit(); cmErr != nil {
		return nil, cmErr
	}
	return cron.Convert2DTO(), nil
}

func (s *provider) CronDelete(ctx context.Context, req *pb.CronDeleteRequest) (*pb.CronDeleteResponse, error) {
	txSession := s.dbClient.NewSession()
	defer txSession.Close()
	if err := txSession.Begin(); err != nil {
		return nil, err
	}

	err := func() error {
		cron, found, err := s.dbClient.GetPipelineCron(req.CronID, mysqlxorm.WithSession(txSession))
		if err != nil {
			return apierrors.ErrDeletePipelineCron.InvalidParameter(err)
		}
		if !found {
			return apierrors.ErrDeletePipelineCron.InvalidParameter(fmt.Errorf("not found"))
		}

		if err := s.dbClient.DeletePipelineCron(cron.ID, mysqlxorm.WithSession(txSession)); err != nil {
			return apierrors.ErrDeletePipelineCron.InternalError(err)
		}

		if err := s.Daemon.DeletePipelineCrond(&cron); err != nil {
			return apierrors.ErrDeletePipelineCron.InternalError(err)
		}

		toEdge := s.EdgePipelineRegister.ShouldDispatchToEdge(cron.PipelineSource.String(), cron.Extra.ClusterName)
		if toEdge {
			bdl, err := s.EdgePipelineRegister.GetEdgeBundleByClusterName(cron.Extra.ClusterName)
			if err != nil {
				return fmt.Errorf("failed to GetEdgeBundleByClusterName error %v", err)
			}

			err = bdl.DeleteCron(cron.ID)
			if err != nil {
				return err
			}
		}
		return nil
	}()

	if err != nil {
		if rbErr := txSession.Rollback(); rbErr != nil {
			return nil, err
		}
		return nil, err
	}
	if cmErr := txSession.Commit(); cmErr != nil {
		return nil, cmErr
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
	toEdge := s.EdgePipelineRegister.ShouldDispatchToEdge(cron.PipelineSource.String(), cron.Extra.ClusterName)
	if toEdge {
		cron.IsEdge = true
		fields = append(fields, spec.PipelineCronIsEdge)
	}

	txSession := s.dbClient.NewSession()
	defer txSession.Close()
	if err := txSession.Begin(); err != nil {
		return nil, err
	}
	err = func() error {
		err = s.dbClient.UpdatePipelineCronWillUseDefault(cron.ID, &cron, fields, mysqlxorm.WithSession(txSession))
		if err != nil {
			return apierrors.ErrUpdatePipelineCron.InternalError(err)
		}

		if err := s.Daemon.AddIntoPipelineCrond(&cron); err != nil {
			return apierrors.ErrUpdatePipelineCron.InternalError(err)
		}

		if toEdge {
			bdl, err := s.EdgePipelineRegister.GetEdgeBundleByClusterName(cron.Extra.ClusterName)
			if err != nil {
				return fmt.Errorf("failed to GetEdgeBundleByClusterName error %v", err)
			}

			err = bdl.UpdateCron(req)
			if err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		if rbErr := txSession.Rollback(); rbErr != nil {
			return nil, err
		}
		return nil, err
	}
	if cmErr := txSession.Commit(); cmErr != nil {
		return nil, cmErr
	}
	return &pb.CronUpdateResponse{}, nil
}

// edge
func (s *provider) EdgeCronCreate(ctx context.Context, req *pb.EdgeCronCreateRequest) (*pb.CronCreateResponse, error) {
	if req.Cron == nil {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("missing cron"))
	}
	if !apistructs.PipelineSource(req.Cron.PipelineSource).Valid() {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("invalid pipelineSource: %s", req.Cron.PipelineSource))
	}
	if req.Cron.PipelineYmlName == "" {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("missing pipelineYmlName"))
	}
	if req.Cron.PipelineYml == "" {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("missing pipelineYml"))
	}

	cron := pbCronToDBCron(req.Cron)
	cron.IsEdge = true

	err := s.dbClient.CreatePipelineCron(cron)
	if err != nil {
		return nil, apierrors.ErrCreatePipelineCron.InternalError(err)
	}

	if err := s.Daemon.AddIntoPipelineCrond(cron); err != nil {
		return nil, apierrors.ErrReloadCrond.InternalError(err)
	}

	return &pb.CronCreateResponse{
		Data: cron.ID,
	}, nil
}

func (s *provider) EdgeCronUpdate(ctx context.Context, req *pb.EdgeCronUpdateRequest) (*pb.CronUpdateResponse, error) {
	if req.Cron == nil {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("missing cron"))
	}

	if req.Cron.ID <= 0 {
		return nil, apierrors.ErrCreatePipelineCron.InvalidParameter(errors.Errorf("missing cron ID"))
	}

	cron := pbCronToDBCron(req.Cron)
	cron.IsEdge = true

	err := s.dbClient.UpdatePipelineCron(cron.ID, cron)
	if err != nil {
		return nil, apierrors.ErrCreatePipelineCron.InternalError(err)
	}

	if err := s.Daemon.AddIntoPipelineCrond(cron); err != nil {
		return nil, apierrors.ErrReloadCrond.InternalError(err)
	}

	return &pb.CronUpdateResponse{}, nil
}

func (s *provider) CronSave(ctx context.Context, req *pb.CronSaveRequest) (*pb.CronSaveResponse, error) {

	cron := pbCronToDBCron(req.Cron)

	txSession := s.dbClient.NewSession()
	defer txSession.Close()
	if err := txSession.Begin(); err != nil {
		return nil, err
	}
	err := func() error {
		err := s.InsertOrUpdatePipelineCron(cron, mysqlxorm.WithSession(txSession))
		if err != nil {
			return err
		}
		req.Cron.ID = cron.ID
		if err := s.Daemon.AddIntoPipelineCrond(cron); err != nil {
			return err
		}
		return nil
	}()
	if err != nil {
		if rbErr := txSession.Rollback(); rbErr != nil {
			return nil, err
		}
		return nil, err
	}
	if cmErr := txSession.Commit(); cmErr != nil {
		return nil, cmErr
	}

	return &pb.CronSaveResponse{
		Data: req.Cron,
	}, nil
}

func (s *provider) CronDisable(ctx context.Context, req *pb.CronDisableRequest) (*pb.CronDisableResponse, error) {

	cron := pbCronToDBCron(req.Cron)

	txSession := s.dbClient.NewSession()
	defer txSession.Close()
	if err := txSession.Begin(); err != nil {
		return nil, err
	}
	err := func() error {
		toEdge := s.EdgePipelineRegister.ShouldDispatchToEdge(cron.PipelineSource.String(), cron.Extra.ClusterName)

		defer func() {
			if err := s.Daemon.AddIntoPipelineCrond(cron); err != nil {
				logrus.Errorf("[alert] add crond failed, err: %v", err)
			}
		}()

		var disable = false
		var updateCron = &db.PipelineCron{}
		var columns = []string{spec.PipelineCronCronExpr, spec.PipelineCronEnable}
		var bdl *bundle.Bundle
		var err error

		if toEdge {
			columns = append(columns, spec.PipelineCronIsEdge)
			updateCron.IsEdge = true
			req.Cron.IsEdge = wrapperspb.Bool(true)
			bdl, err = s.EdgePipelineRegister.GetEdgeBundleByClusterName(cron.Extra.ClusterName)
			if err != nil {
				return fmt.Errorf("failed to GetEdgeBundleByClusterName error %v", err)
			}
		}

		queryV1 := &db.PipelineCron{
			ApplicationID:   cron.ApplicationID,
			Branch:          cron.Branch,
			PipelineYmlName: cron.PipelineYmlName,
		}
		v1Exist, err := txSession.Get(queryV1)
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
			err := s.dbClient.UpdatePipelineCronWillUseDefault(updateCron.ID, updateCron, columns, mysqlxorm.WithSession(txSession))
			if err != nil {
				return err
			}
			req.Cron.ID = updateCron.ID

			if toEdge {
				err := bdl.CronDisable(req)
				if err != nil {
					return err
				}
			}
			return nil
		}

		queryV2 := &spec.PipelineCron{
			PipelineSource:  cron.PipelineSource,
			PipelineYmlName: cron.PipelineYmlName,
		}
		v2Exist, err := txSession.Get(queryV2)
		if err != nil {
			return err
		}
		if v2Exist {
			updateCron.Enable = &disable
			updateCron.ID = queryV2.ID
			updateCron.CronExpr = cron.CronExpr
			err := s.dbClient.UpdatePipelineCronWillUseDefault(updateCron.ID, updateCron, columns, mysqlxorm.WithSession(txSession))
			if err != nil {
				return err
			}
			req.Cron.ID = updateCron.ID

			if toEdge {
				err := bdl.CronDisable(req)
				if err != nil {
					return err
				}
			}
			return nil
		}

		return nil
	}()

	if err != nil {
		if rbErr := txSession.Rollback(); rbErr != nil {
			return nil, err
		}
		return nil, err
	}
	if cmErr := txSession.Commit(); cmErr != nil {
		return nil, cmErr
	}

	return &pb.CronDisableResponse{
		Data: req.Cron,
	}, nil
}

func pbCronToDBCron(pbCron *common.Cron) *db.PipelineCron {
	cron := db.PipelineCron{
		ID:                   pbCron.ID,
		ApplicationID:        pbCron.ApplicationID,
		Branch:               pbCron.Branch,
		PipelineSource:       apistructs.PipelineSource(pbCron.PipelineSource),
		PipelineYmlName:      pbCron.PipelineYmlName,
		CronExpr:             pbCron.CronExpr,
		Enable:               &[]bool{pbCron.Enable.Value}[0],
		PipelineDefinitionID: pbCron.PipelineDefinitionID,
		IsEdge: func() bool {
			if pbCron.IsEdge == nil {
				return false
			}
			return pbCron.IsEdge.Value
		}(),
	}

	var extra = db.PipelineCronExtra{
		PipelineYml:            pbCron.CronExtra.PipelineYml,
		ClusterName:            pbCron.CronExtra.ClusterName,
		FilterLabels:           pbCron.CronExtra.FilterLabels,
		NormalLabels:           pbCron.CronExtra.NormalLabels,
		Envs:                   pbCron.CronExtra.Envs,
		Version:                pbCron.CronExtra.Version,
		ConfigManageNamespaces: pbCron.CronExtra.ConfigManageNamespaces,
		IncomingSecrets:        pbCron.CronExtra.IncomingSecrets,
		CronStartFrom:          &[]time.Time{pbCron.CronExtra.CronStartFrom.AsTime()}[0],
		LastCompensateAt:       &[]time.Time{pbCron.CronExtra.LastCompensateAt.AsTime()}[0],
	}

	if pbCron.CronExtra.Compensator != nil {
		extra.Compensator = &apistructs.CronCompensator{
			Enable:               pbCron.CronExtra.Compensator.Enable.Value,
			LatestFirst:          pbCron.CronExtra.Compensator.LatestFirst.Value,
			StopIfLatterExecuted: pbCron.CronExtra.Compensator.StopIfLatterExecuted.Value,
		}
	}
	cron.Extra = extra
	cron.ID = pbCron.ID
	return &cron
}
