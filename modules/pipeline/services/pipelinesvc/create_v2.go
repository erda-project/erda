// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package pipelinesvc

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func (s *PipelineSvc) CreateV2(req *apistructs.PipelineCreateRequestV2) (*spec.Pipeline, error) {
	// validate
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}
	// set default
	setDefault(req)

	p, err := s.makePipelineFromRequestV2(req)
	if err != nil {
		return nil, err
	}

	if err := s.createPipelineGraph(p); err != nil {
		logrus.Errorf("failed to create pipeline graph, pipelineID: %d, err: %v", p.ID, err)
		return nil, err
	}

	// 立即执行一次
	if req.AutoRunAtOnce {
		_p, err := s.RunPipeline(&apistructs.PipelineRunRequest{
			PipelineID:        p.ID,
			ForceRun:          req.ForceRun,
			IdentityInfo:      req.IdentityInfo,
			PipelineRunParams: req.RunParams,
		})
		if err != nil {
			logrus.Errorf("failed to run pipeline, pipelineID: %d, err: %v", p.ID, err)
			return nil, err
		}
		p = _p
	}

	// 立即开始定时
	if req.AutoStartCron {
		if p.CronID != nil {
			if _, err := s.pipelineCronSvc.Start(*p.CronID); err != nil {
				logrus.Errorf("failed to start cron, pipelineID: %d, cronID: %d, err: %v", p.ID, *p.CronID, err)
				return nil, err
			}
		}
	}

	return p, nil
}

const (
	maxSqlIndexLength = 191
)

// validateCreateRequest validate pipelineCreateRequestV2
func (s *PipelineSvc) validateCreateRequest(req *apistructs.PipelineCreateRequestV2) error {
	if req == nil {
		return apierrors.ErrCreatePipeline.MissingParameter("request")
	}
	// +required
	if req.PipelineYml == "" {
		return apierrors.ErrCreatePipeline.MissingParameter("pipelineYml")
	}
	// +required
	if req.ClusterName == "" {
		return apierrors.ErrCreatePipeline.MissingParameter("clusterName")
	}
	// +optional
	if req.PipelineYmlName == "" {
		req.PipelineYmlName = apistructs.DefaultPipelineYmlName
	}
	// +required
	if req.PipelineSource == "" {
		return apierrors.ErrCreatePipeline.MissingParameter("pipelineSource")
	}
	if !req.PipelineSource.Valid() {
		return apierrors.ErrCreatePipeline.InvalidParameter(errors.Errorf("source: %s", req.PipelineSource))
	}
	// identity
	if req.UserID == "" && req.InternalClient == "" {
		return apierrors.ErrCreatePipeline.MissingParameter("identity")
	}
	// filterLabels
	// if label key or value is too long, it will be moved to NormalLabels automatically.
	if req.NormalLabels == nil {
		req.NormalLabels = make(map[string]string)
	}
	for k, v := range req.Labels {
		if len(k) > maxSqlIndexLength || len(v) > maxSqlIndexLength {
			logrus.Warnf("filterLabel key or value is too long, move to normalLabels automatically, key: %s, value: %s", k, v)
			req.NormalLabels[k] = v
			delete(req.Labels, k)
		}
	}
	// bind queue
	_, err := s.validateQueueFromLabels(req)
	if err != nil {
		return apierrors.ErrCreatePipeline.InvalidParameter(err)
	}
	return nil
}

// setDefault set default value for PipelineCreateRequestV2
func setDefault(req *apistructs.PipelineCreateRequestV2) {
	if req.PipelineYmlName == "" {
		req.PipelineYmlName = apistructs.DefaultPipelineYmlName
	}
}

func logCompatibleFailed(key, value string, err error) {
	logrus.Errorf("compatible from labels failed, key: %s, value: %s, err: %v", key, value, err)
}

func (s *PipelineSvc) makePipelineFromRequestV2(req *apistructs.PipelineCreateRequestV2) (*spec.Pipeline, error) {
	p := &spec.Pipeline{}

	// 解析 pipeline yml 文件，生成最终 pipeline yml 文件
	// 只解析最外层，获取 storage 和 cron 信息
	pipelineYml, err := pipelineyml.New([]byte(req.PipelineYml), pipelineyml.WithEnvs(req.Envs))
	if err != nil {
		return nil, apierrors.ErrParsePipelineYml.InternalError(err)
	}

	p.PipelineYml = req.PipelineYml
	p.PipelineYmlName = req.PipelineYmlName
	p.PipelineSource = req.PipelineSource
	p.ClusterName = req.ClusterName

	// labels
	p.NormalLabels = req.NormalLabels
	if p.NormalLabels == nil {
		p.NormalLabels = make(map[string]string)
	}
	p.Labels = req.Labels
	if p.Labels == nil {
		p.Labels = make(map[string]string)
	}

	// envs
	p.Snapshot.Envs = req.Envs
	p.Snapshot.RunPipelineParams = req.RunParams.ToPipelineRunParamsWithValue()

	// status
	p.Status = apistructs.PipelineStatusAnalyzed

	// identity
	if req.UserID != "" {
		p.Extra.SubmitUser = s.tryGetUser(req.UserID)
	}
	p.Extra.InternalClient = req.InternalClient
	p.Snapshot.IdentityInfo = req.IdentityInfo

	// 挂载配置
	p.Extra.StorageConfig.EnableNFS = true
	storageConfig := pipelineYml.Spec().Storage
	if storageConfig != nil && storageConfig.Context == "local" {
		p.Extra.StorageConfig.EnableLocal = true
	}
	// 是否全局配置开启流水线挂载
	if conf.DisablePipelineVolume() {
		p.Extra.StorageConfig.EnableNFS = false
		p.Extra.StorageConfig.EnableLocal = false
	}

	// auto run
	p.Extra.IsAutoRun = req.AutoRun

	version, err := pipelineyml.GetVersion([]byte(p.PipelineYml))
	if err != nil {
		return nil, apierrors.ErrParsePipelineYml.InvalidParameter(errors.Errorf("version (%v)", err))
	}
	p.Extra.Version = version

	p.CostTimeSec = -1

	// 解析 labels，填充表字段
	labels := p.MergeLabels()

	// workspace
	p.Extra.DiceWorkspace = apistructs.DiceWorkspace(labels[apistructs.LabelDiceWorkspace])

	// vcs
	if v, ok := labels[apistructs.LabelCommitDetail]; ok {
		var detail apistructs.CommitDetail
		if err := json.Unmarshal([]byte(v), &detail); err != nil {
			logCompatibleFailed(apistructs.LabelCommitDetail, v, err)
		}
		p.CommitDetail = detail
	}
	if v, ok := labels[apistructs.LabelCommit]; ok {
		p.CommitDetail.CommitID = v
	}

	// pipelineYmlSource
	p.Extra.PipelineYmlSource = apistructs.PipelineYmlSourceContent
	if v, ok := labels[apistructs.LabelPipelineYmlSource]; ok {
		if !apistructs.PipelineYmlSource(v).Valid() {
			logCompatibleFailed(apistructs.LabelPipelineYmlSource, v, nil)
		}
		p.Extra.PipelineYmlSource = apistructs.PipelineYmlSource(v)
	}

	// pipelineType
	if v, ok := labels[apistructs.LabelPipelineType]; ok {
		if !apistructs.PipelineType(v).Valid() {
			logCompatibleFailed(apistructs.LabelPipelineType, v, nil)
		}
		p.Type = apistructs.PipelineType(v)
	}

	// cronTriggerTime
	if v, ok := labels[apistructs.LabelPipelineCronTriggerTime]; ok {
		nano, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, apierrors.ErrCreatePipeline.InvalidParameter(err)
		}
		cronTriggerTime := time.Unix(0, nano)
		p.Extra.CronTriggerTime = &cronTriggerTime
	}

	// configManage
	p.Extra.ConfigManageNamespaces = req.ConfigManageNamespaces

	// cron
	p.Extra.CronExpr = pipelineYml.Spec().Cron
	if v, ok := labels[apistructs.LabelPipelineCronID]; ok {
		cronID, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, apierrors.ErrCreatePipeline.InvalidParameter(err)
		}
		pc, err := s.dbClient.GetPipelineCron(cronID)
		if err != nil {
			return nil, apierrors.ErrGetPipelineCron.InvalidParameter(err)
		}
		p.CronID = &pc.ID
		p.Extra.CronExpr = pc.CronExpr
	}

	// triggerMode
	if v, ok := labels[apistructs.LabelPipelineTriggerMode]; ok {
		if !apistructs.PipelineTriggerMode(v).Valid() {
			logCompatibleFailed(apistructs.LabelPipelineTriggerMode, v, nil)
		}
		p.TriggerMode = apistructs.PipelineTriggerMode(v)
	}

	// progress
	p.Progress = -1

	// gc
	p.Extra.GC = req.GC

	if err := s.UpdatePipelineCron(p, req.CronStartFrom, req.ConfigManageNamespaces, pipelineYml.Spec().CronCompensator); err != nil {
		return nil, apierrors.ErrCreatePipeline.InternalError(err)
	}

	// defined outputs
	for _, output := range pipelineYml.Spec().Outputs {
		p.Extra.DefinedOutputs = append(p.Extra.DefinedOutputs,
			apistructs.PipelineOutput{
				Name: output.Name,
				Desc: output.Desc,
				Ref:  output.Ref,
			})
	}

	// queue
	if req.BindQueue != nil {
		p.Extra.QueueID = req.BindQueue.ID
	}

	return p, nil
}

// 非定时触发的，如果有定时配置，需要插入或更新 pipeline_crons enable 配置
// 不管是定时还是非定时，只要定时配置是空的，就将pipeline_crons disable
func (s *PipelineSvc) UpdatePipelineCron(p *spec.Pipeline, cronStartFrom *time.Time, configManageNamespaces []string, cronCompensator *pipelineyml.CronCompensator) error {

	var cron *spec.PipelineCron

	//是定时类型的流水线，切定时的表达式不为空，更新cron的配置
	if p.TriggerMode != apistructs.PipelineTriggerModeCron && p.Extra.CronExpr != "" {

		cron = constructPipelineCron(p, cronStartFrom, configManageNamespaces, cronCompensator)

		if err := s.dbClient.InsertOrUpdatePipelineCron(cron); err != nil {
			return apierrors.ErrUpdatePipelineCron.InternalError(err)
		}
		p.CronID = &cron.ID

		if err := s.crondSvc.DistributedReloadCrond(); err != nil {
			logrus.Errorf("[alert] distributed reload crond failed, err: %v", err)
		}
	}
	//cron表达式为空，就需要关闭定时
	if p.Extra.CronExpr == "" {

		cron = constructPipelineCron(p, cronStartFrom, configManageNamespaces, cronCompensator)
		if err := s.dbClient.DisablePipelineCron(cron); err != nil {
			return apierrors.ErrUpdatePipelineCron.InternalError(err)
		}
		p.CronID = nil

		if err := s.crondSvc.DistributedReloadCrond(); err != nil {
			logrus.Errorf("[alert] distributed reload crond failed, err: %v", err)
		}
	}

	return nil
}

func constructPipelineCron(p *spec.Pipeline, cronStartFrom *time.Time, configManageNamespaces []string, cronCompensator *pipelineyml.CronCompensator) *spec.PipelineCron {
	appID, _ := strconv.ParseUint(p.Labels[apistructs.LabelAppID], 10, 64)
	var compensator *apistructs.CronCompensator
	if cronCompensator != nil {
		compensator = &apistructs.CronCompensator{}
		compensator.Enable = cronCompensator.Enable
		compensator.LatestFirst = cronCompensator.LatestFirst
		compensator.StopIfLatterExecuted = cronCompensator.StopIfLatterExecuted
	}
	cron := &spec.PipelineCron{
		ApplicationID:   appID,
		Branch:          p.Labels[apistructs.LabelBranch],
		PipelineSource:  p.PipelineSource,
		PipelineYmlName: p.PipelineYmlName,
		CronExpr:        p.Extra.CronExpr,
		Enable:          &[]bool{false}[0],
		Extra: spec.PipelineCronExtra{
			PipelineYml:            p.PipelineYml,
			ClusterName:            p.ClusterName,
			FilterLabels:           p.Labels,
			NormalLabels:           p.GenerateNormalLabelsForCreateV2(),
			Envs:                   p.Snapshot.Envs,
			ConfigManageNamespaces: configManageNamespaces,
			CronStartFrom:          cronStartFrom,
			Version:                "v2",
			Compensator:            compensator,
			LastCompensateAt:       nil,
		},
	}

	return cron
}
