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
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"xorm.io/xorm"

	commonpb "github.com/erda-project/erda-proto-go/common/pb"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	queuepb "github.com/erda-project/erda-proto-go/core/pipeline/queue/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/conf"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/events"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/action_info"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/container_provider"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/cron/crontypes"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

const (
	maxSqlIndexLength = 191
)

func (s *pipelineService) PipelineCreateV2(ctx context.Context, req *pb.PipelineCreateRequestV2) (*pb.PipelineCreateResponse, error) {
	if req.AutoRun {
		req.AutoRunAtOnce = true
	}
	identityInfo := apis.GetIdentityInfo(ctx)
	if req.UserID == "" && identityInfo != nil {
		req.UserID = identityInfo.UserID
	}
	if req.InternalClient == "" && identityInfo != nil {
		req.InternalClient = identityInfo.InternalClient
	}

	// bind queue is internal use
	req.BindQueue = nil

	canProxy := s.edgeRegister.CanProxyToEdge(apistructs.PipelineSource(req.PipelineSource), req.ClusterName)

	if canProxy {
		s.p.Log.Infof("proxy create pipeline to edge, source: %s, yamlName: %s", req.PipelineSource, req.PipelineYmlName)
		edgePipeline, err := s.proxyCreatePipelineRequestToEdge(ctx, req)
		if err != nil {
			return nil, err
		}
		return &pb.PipelineCreateResponse{Data: edgePipeline}, nil
	}
	p, err := s.CreateV2(ctx, req)
	if err != nil {
		return nil, err
	}
	// report
	if s.edgeRegister.IsEdge() {
		s.edgeReporter.TriggerOncePipelineReport(p.ID)
	}
	pipelineDto := s.ConvertPipeline(p)

	return &pb.PipelineCreateResponse{
		Data: pipelineDto,
	}, nil
}

func (s *pipelineService) proxyCreatePipelineRequestToEdge(ctx context.Context, req *pb.PipelineCreateRequestV2) (*basepb.PipelineDTO, error) {
	// handle at edge side
	edgeBundle, err := s.edgeRegister.GetEdgeBundleByClusterName(req.ClusterName)
	if err != nil {
		return nil, err
	}
	p, err := edgeBundle.CreatePipeline(req)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *pipelineService) CreateV2(ctx context.Context, req *pb.PipelineCreateRequestV2) (*spec.Pipeline, error) {
	// validate
	if err := s.ValidateCreateRequest(req); err != nil {
		return nil, err
	}

	// set default
	setDefault(req)

	p, err := s.MakePipelineFromRequestV2(req)
	if err != nil {
		return nil, err
	}

	var stages []spec.PipelineStage
	if stages, err = s.CreatePipelineGraph(p); err != nil {
		s.p.Log.Errorf("failed to create pipeline graph, pipelineID: %d, err: %v", p.ID, err)
		return nil, err
	}

	// PreCheck
	_ = s.PreCheck(p, stages, p.GetUserID(), req.AutoRunAtOnce)

	// do it once
	if req.AutoRunAtOnce {
		_p, err := s.run.RunOnePipeline(ctx, &pb.PipelineRunRequest{
			PipelineID:        p.ID,
			ForceRun:          req.ForceRun,
			UserID:            req.UserID,
			InternalClient:    req.InternalClient,
			PipelineRunParams: req.RunParams,
			Secrets:           req.Secrets,
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
			if _, err := s.cronSvc.CronStart(context.Background(), &cronpb.CronStartRequest{
				CronID: *p.CronID,
			}); err != nil {
				logrus.Errorf("failed to start cron, pipelineID: %d, cronID: %d, err: %v", p.ID, *p.CronID, err)
				return nil, err
			}
		}
	}
	return p, nil
}

// CreatePipelineGraph recursively create pipeline graph.
func (s *pipelineService) CreatePipelineGraph(p *spec.Pipeline) (newStages []spec.PipelineStage, err error) {
	// parse yml
	pipelineYml, err := pipelineyml.New(
		[]byte(p.PipelineYml),
	)
	if err != nil {
		return nil, apierrors.ErrParsePipelineYml.InternalError(err)
	}

	// init pipeline gc setting
	p.EnsureGC()

	// only create pipeline and stages, tasks waiting pipeline run
	var stages []*spec.PipelineStage
	_, err = s.dbClient.Transaction(func(session *xorm.Session) (interface{}, error) {
		// create pipeline
		if err := s.createPipelineAndCheckNotEndStatus(p, session); err != nil {
			return nil, err
		}
		// create pipeline stages
		stages, err = s.createPipelineGraphStage(p, pipelineYml, dbclient.WithTxSession(session))
		if err != nil {
			return nil, err
		}

		// calculate pipeline applied resource after all snippetTask created
		pipelineAppliedResources, err := s.resource.CalculatePipelineResources(pipelineYml, p)
		if err != nil {
			return nil, apierrors.ErrCreatePipelineGraph.InternalError(
				fmt.Errorf("failed to search pipeline action resrouces, err: %v", err))
		}
		if pipelineAppliedResources != nil {
			p.Snapshot.AppliedResources = *pipelineAppliedResources
		}
		if err := s.dbClient.UpdatePipelineExtraSnapshot(p.ID, p.Snapshot, dbclient.WithTxSession(session)); err != nil {
			return nil, apierrors.ErrCreatePipelineGraph.InternalError(
				fmt.Errorf("failed to update pipeline snapshot for applied resources, err: %v", err))
		}
		return nil, nil
	})
	if err != nil {
		return nil, apierrors.ErrCreatePipelineGraph.InternalError(err)
	}

	// cover stages
	for _, stage := range stages {
		newStages = append(newStages, *stage)
	}

	// events
	events.EmitPipelineInstanceEvent(p, p.GetSubmitUserID())
	return newStages, nil
}

func (s *pipelineService) createPipelineAndCheckNotEndStatus(p *spec.Pipeline, session *xorm.Session) error {
	// Check whether the parent pipeline has an end state
	for _, parentPipelineID := range p.Extra.SnippetChain {
		parentPipeline, _, err := s.dbClient.GetPipelineBase(parentPipelineID, dbclient.WithTxSession(session))
		if err != nil {
			logrus.Errorf("check whether the parent pipeline has an end state, error %v", err)
			continue
		}
		if parentPipeline.Status.IsEndStatus() {
			return fmt.Errorf("parent pipeline was end status")
		}
	}

	// create pipeline
	if s.edgeRegister.IsEdge() {
		p.IsEdge = true
	}
	if err := s.dbClient.CreatePipeline(p, dbclient.WithTxSession(session)); err != nil {
		return apierrors.ErrCreatePipeline.InternalError(err)
	}
	return nil
}

// traverse the stage of yml and save it to the database
func (s *pipelineService) createPipelineGraphStage(p *spec.Pipeline, pipelineYml *pipelineyml.PipelineYml, ops ...dbclient.SessionOption) (stages []*spec.PipelineStage, err error) {
	var dbStages []*spec.PipelineStage

	for si := range pipelineYml.Spec().Stages {
		ps := &spec.PipelineStage{
			PipelineID:  p.ID,
			Name:        "",
			Status:      apistructs.PipelineStatusAnalyzed,
			CostTimeSec: -1,
			Extra:       spec.PipelineStageExtra{StageOrder: si},
		}
		if err := s.dbClient.CreatePipelineStage(ps, ops...); err != nil {
			return nil, apierrors.ErrCreatePipelineGraph.InternalError(err)
		}
		dbStages = append(dbStages, ps)
	}
	return dbStages, nil
}

// setDefault set default value for PipelineCreateRequestV2
func setDefault(req *pb.PipelineCreateRequestV2) {
	if req.PipelineYmlName == "" {
		req.PipelineYmlName = apistructs.DefaultPipelineYmlName
	}
}

// ValidateCreateRequest validate pipelineCreateRequestV2
func (s *pipelineService) ValidateCreateRequest(req *pb.PipelineCreateRequestV2) error {
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
	if !apistructs.PipelineSource(req.PipelineSource).Valid() {
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

func (s *pipelineService) validateQueueFromLabels(req *pb.PipelineCreateRequestV2) (*queuepb.Queue, error) {
	var foundBindQueueID bool
	var bindQueueIDStr string
	for k, v := range req.Labels {
		if k == apistructs.LabelBindPipelineQueueID {
			foundBindQueueID = true
			bindQueueIDStr = v
			break
		}
	}
	if !foundBindQueueID {
		return nil, nil
	}
	// parse queue id
	queueID, err := strconv.ParseUint(bindQueueIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bindQueueID: %s, err: %v", bindQueueIDStr, err)
	}
	// query queue
	queueRes, err := s.queueManage.GetQueue(context.Background(), &queuepb.QueueGetRequest{QueueID: queueID})
	if err != nil {
		return nil, err
	}
	// check queue is matchable
	if err := checkQueueValidateWithPipelineCreateReq(req, queueRes.Data); err != nil {
		return nil, err
	}
	req.BindQueue = queueRes.Data

	return queueRes.Data, nil
}

func checkQueueValidateWithPipelineCreateReq(req *pb.PipelineCreateRequestV2, queue *queuepb.Queue) error {
	// pipeline source
	if queue.PipelineSource != req.PipelineSource {
		return fmt.Errorf("invalid queue: pipeline source not match: %s(req) vs %s(queue)", req.PipelineSource, queue.PipelineSource)
	}
	// cluster name
	if queue.ClusterName != req.ClusterName {
		return fmt.Errorf("invalid queue: cluster name not match: %s(req) vs %s(queue)", req.ClusterName, queue.ClusterName)
	}

	return nil
}

func (s *pipelineService) MakePipelineFromRequestV2(req *pb.PipelineCreateRequestV2) (*spec.Pipeline, error) {
	p := &spec.Pipeline{}

	// 解析 pipeline yml 文件，生成最终 pipeline yml 文件
	// 只解析最外层，获取 storage 和 cron 信息
	pipelineYml, err := pipelineyml.New([]byte(req.PipelineYml), pipelineyml.WithEnvs(req.Envs))
	if err != nil {
		return nil, apierrors.ErrParsePipelineYml.InternalError(err)
	}

	p.PipelineYml = req.PipelineYml
	p.PipelineYmlName = req.PipelineYmlName
	p.PipelineSource = apistructs.PipelineSource(req.PipelineSource)
	p.ClusterName = req.ClusterName
	p.PipelineDefinitionID = req.DefinitionID
	// labels
	p.NormalLabels = req.NormalLabels
	if p.NormalLabels == nil {
		p.NormalLabels = make(map[string]string)
	}
	p.Labels = req.Labels
	if p.Labels == nil {
		p.Labels = make(map[string]string)
	}
	p.Labels[apistructs.LabelCreateUserID] = req.UserID

	// envs
	p.Snapshot.Envs = req.Envs
	p.Snapshot.RunPipelineParams = s.ToPipelineRunParamsWithValue(req.RunParams)

	// status
	p.Status = apistructs.PipelineStatusAnalyzed

	// identity
	if req.UserID != "" {
		p.Extra.SubmitUser = s.user.TryGetUser(context.Background(), req.UserID)
	}
	p.Extra.OwnerUser = req.OwnerUser
	p.Extra.InternalClient = req.InternalClient
	if req.UserID != "" {
		p.Snapshot.IdentityInfo = commonpb.IdentityInfo{
			UserID:         req.UserID,
			InternalClient: req.InternalClient,
		}
	}
	if p.GetOwnerUserID() != "" {
		p.Labels[apistructs.LabelOwnerUserID] = p.GetOwnerUserID()
	}

	// namespace
	// if upper layer customize namespace, use custom namespace
	// default namespace is pipeline controlled
	if req.Namespace != "" {
		p.Extra.Namespace = req.Namespace
		p.Extra.NotPipelineControlledNs = true
	}

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

	// org
	if p.GetOrgName() == "" {
		p.NormalLabels[apistructs.LabelOrgName] = s.tryGetOrgName(p)
	}

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

	// container instance provider
	extensionItems := make([]string, 0)
	for _, stage := range pipelineYml.Spec().Stages {
		for _, actionMap := range stage.Actions {
			for _, action := range actionMap {
				if action.Type.IsSnippet() {
					continue
				}
				extensionItems = append(extensionItems, s.actionMgr.MakeActionTypeVersion(action))
			}
		}
	}
	_, extensions, err := s.actionMgr.SearchActions(extensionItems)
	if err != nil {
		return nil, apierrors.ErrCreatePipeline.InternalError(err)
	}
	p.Extra.ContainerInstanceProvider = container_provider.ConstructContainerProvider(container_provider.WithLabels(labels),
		container_provider.WithExtensions(extensions))

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

	// secrets
	p.Extra.IncomingSecrets = req.Secrets

	// cron
	p.Extra.CronExpr = pipelineYml.Spec().Cron
	if v, ok := labels[apistructs.LabelPipelineCronID]; ok {
		cronID, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return nil, apierrors.ErrCreatePipeline.InvalidParameter(err)
		}
		pc, err := s.cronSvc.CronGet(context.Background(), &cronpb.CronGetRequest{
			CronID: cronID,
		})
		if err != nil {
			return nil, apierrors.ErrGetPipelineCron.InvalidParameter(err)
		}
		if pc.Data == nil {
			return nil, apierrors.ErrNotFoundPipelineCron.InvalidParameter(crontypes.ErrCronNotFound)
		}
		p.CronID = &pc.Data.ID
		p.Extra.CronExpr = pc.Data.CronExpr
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

	// breakpoint
	p.Extra.Breakpoint = pipelineYml.Spec().Breakpoint

	// gc
	if req.GC != nil {
		p.Extra.GC = *req.GC
	} else {
		p.Extra.GC = basepb.PipelineGC{}
	}
	initializePipelineGC(&p.Extra.GC)

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
		customPriority := req.BindQueue.Priority
		customPriorityStr, ok := p.MergeLabels()[apistructs.LabelBindPipelineQueueCustomPriority]
		if ok {
			_customPriority, err := strconv.ParseInt(customPriorityStr, 10, 64)
			if err == nil {
				customPriority = _customPriority
			}
		}
		p.Extra.QueueInfo = &spec.QueueInfo{
			QueueID:          req.BindQueue.ID,
			CustomPriority:   customPriority,
			EnqueueCondition: apistructs.EnqueueConditionType(p.MergeLabels()[apistructs.LabelBindPipelineQueueEnqueueCondition]),
		}
	}

	return p, nil
}

// 非定时触发的，如果有定时配置，需要插入或更新 pipeline_crons enable 配置
// 不管是定时还是非定时，只要定时配置是空的，就将pipeline_crons disable
func (s *pipelineService) UpdatePipelineCron(p *spec.Pipeline, cronStartFrom *timestamppb.Timestamp, configManageNamespaces []string, cronCompensator *pipelineyml.CronCompensator) error {
	if p.TriggerMode == apistructs.PipelineTriggerModeCron {
		return nil
	}

	var createRequest = constructToCreateCronRequest(p, cronStartFrom, configManageNamespaces)
	result, err := s.cronSvc.CronCreate(context.Background(), createRequest)
	if err != nil {
		return apierrors.ErrUpdatePipelineCron.InternalError(err)
	}
	// todo CronCreate should be simple, do not contains disable logic.
	//  Add an interface method HandleCron to merge all logic for easy use.
	// When cron create cron express is empty, cron create will execute disable logic, if not find cron by source and ymlName the ID of result may be 0
	if result.Data.ID > 0 {
		p.CronID = &result.Data.ID
	}

	// report
	if s.edgeRegister != nil {
		if s.edgeRegister.IsEdge() && p.CronID != nil {
			s.edgeReporter.TriggerOnceCronReport(*p.CronID)
		}
	}
	return nil
}

func logCompatibleFailed(key, value string, err error) {
	logrus.Errorf("compatible from labels failed, key: %s, value: %s, err: %v", key, value, err)
}

func constructToCreateCronRequest(p *spec.Pipeline, cronStartFrom *timestamppb.Timestamp, configManageNamespaces []string) *cronpb.CronCreateRequest {
	createReq := &cronpb.CronCreateRequest{
		PipelineSource:         p.PipelineSource.String(),
		PipelineYmlName:        p.PipelineYmlName,
		CronExpr:               p.Extra.CronExpr,
		Enable:                 wrapperspb.Bool(false),
		PipelineYml:            p.PipelineYml,
		ClusterName:            p.ClusterName,
		FilterLabels:           p.Labels,
		NormalLabels:           p.GenerateNormalLabelsForCreateV2(),
		Envs:                   p.Snapshot.Envs,
		ConfigManageNamespaces: configManageNamespaces,
		CronStartFrom:          cronStartFrom,
		IncomingSecrets:        p.Extra.IncomingSecrets,
		PipelineDefinitionID:   p.PipelineDefinitionID,
	}
	return createReq
}

func (s *pipelineService) ToPipelineRunParamsWithValue(params []*basepb.PipelineRunParam) []apistructs.PipelineRunParamWithValue {
	var result []apistructs.PipelineRunParamWithValue
	for _, rp := range params {
		result = append(result, apistructs.PipelineRunParamWithValue{PipelineRunParam: apistructs.PipelineRunParam{Name: rp.Name, Value: rp.Value.AsInterface()}})
	}
	return result
}

// replace the tasks parsed by yml and tasks in the database with the same name
func (s *pipelineService) MergePipelineYmlTasks(pipelineYml *pipelineyml.PipelineYml, dbTasks []spec.PipelineTask, p *spec.Pipeline, dbStages []spec.PipelineStage, passedDataWhenCreate *action_info.PassedDataWhenCreate) (mergeTasks []spec.PipelineTask, err error) {
	// loop yml actions to make actionTasks
	actionTasks := s.GetYmlActionTasks(pipelineYml, p, dbStages, passedDataWhenCreate)

	// determine whether the task status was disabled or Paused according to the TaskOperates of the pipeline
	var operateActionTasks []spec.PipelineTask
	for _, actionTask := range actionTasks {
		operateTask, err := s.OperateTask(p, &actionTask)
		if err != nil {
			return nil, apierrors.ErrListPipelineTasks.InternalError(err)
		}
		operateActionTasks = append(operateActionTasks, *operateTask)
	}

	// combine the task converted from yml with the task in the database
	return ymlTasksMergeDBTasks(operateActionTasks, dbTasks), nil
}

// combine the task converted from yml with the task in the database
func ymlTasksMergeDBTasks(actionTasks []spec.PipelineTask, dbTasks []spec.PipelineTask) []spec.PipelineTask {
	var mergeTasks []spec.PipelineTask
	for actionIndex := range actionTasks {
		var actionTask = actionTasks[actionIndex]
		// actionTask the same dbTask pipelineID,stagesID,type and name replace with dbTask
		var mergeTask *spec.PipelineTask
		for index := range dbTasks {
			var dbTask = dbTasks[index]
			if actionTask.PipelineID != dbTask.PipelineID || actionTask.StageID != dbTask.StageID {
				continue
			}
			if actionTask.Type != dbTask.Type || actionTask.Name != dbTask.Name {
				continue
			}
			mergeTask = &dbTask
		}
		if mergeTask == nil {
			mergeTask = &actionTask
		}

		mergeTasks = append(mergeTasks, *mergeTask)
	}
	return mergeTasks
}

// determine whether the task status is disabled according to the TaskOperates of the pipeline and task extra disable field
func (s *pipelineService) OperateTask(p *spec.Pipeline, task *spec.PipelineTask) (*spec.PipelineTask, error) {
	for _, taskOp := range p.Extra.TaskOperates {
		// the name of the disabled task matches the task name
		if taskOp.TaskAlias != task.Name {
			continue
		}
		// task have id mean can not disable
		if task.ID > 0 {
			continue
		}

		var opAction OperateAction
		wrapError := func(err error) error {
			return errors.Wrapf(err, "failed to operate pipeline, task [%v], action [%s]", taskOp.TaskAlias, opAction)
		}
		// disable
		if taskOp.Disable != nil {
			if *taskOp.Disable {
				opAction = OpDisableTask
				if !(task.Status == apistructs.PipelineStatusAnalyzed || task.Status == apistructs.PipelineStatusPaused) {
					return nil, wrapError(errors.Errorf("invalid status [%v]", task.Status))
				}
				task.Status = apistructs.PipelineStatusDisabled
			} else {
				opAction = OpEnableTask
				task.Status = apistructs.PipelineStatusAnalyzed
			}
			// needUpdatePipelineCtx = true
		}

		// pause: task cannot be modified after starting execution
		if taskOp.Pause != nil {
			if *taskOp.Pause {
				opAction = OpPauseTask
				if !task.Status.CanPause() {
					return nil, wrapError(errors.Errorf("status [%s]", task.Status))
				}
				task.Status = apistructs.PipelineStatusPaused
			} else {
				opAction = OpUnpauseTask
				if !task.Status.CanUnpause() {
					return nil, wrapError(errors.Errorf("status [%s]", task.Status))
				}
				// Determine the stage status of the current node:
				// 1. If it is Born, it means that the stage can be pushed by the thruster, then task.status = Born
				// 2. Otherwise, it means that the stage has been executed and will not be advanced again, so task.status = Mark
				stage, err := s.dbClient.GetPipelineStage(task.StageID)
				if err != nil {
					return nil, err
				}
				if stage.Status == apistructs.PipelineStatusBorn {
					task.Status = apistructs.PipelineStatusBorn
				} else {
					task.Status = apistructs.PipelineStatusMark
				}
			}
			task.Extra.Pause = *taskOp.Pause
		}
	}
	return task, nil
}

// GetYmlActionTasks generate task array according to yml structure
func (s *pipelineService) GetYmlActionTasks(pipelineYml *pipelineyml.PipelineYml, p *spec.Pipeline, dbStages []spec.PipelineStage, passedDataWhenCreate *action_info.PassedDataWhenCreate) []spec.PipelineTask {
	if pipelineYml == nil || p == nil || len(dbStages) <= 0 {
		return nil
	}

	// loop yml actions to make actionTasks
	var actionTasks []spec.PipelineTask
	pipelineYml.Spec().LoopStagesActions(func(stageIndex int, action *pipelineyml.Action) {
		var task *spec.PipelineTask
		if action.Type.IsSnippet() {
			task = s.makeSnippetPipelineTask(p, &dbStages[stageIndex], action)
		} else {
			task = s.makeNormalPipelineTask(p, &dbStages[stageIndex], action, passedDataWhenCreate)
		}
		actionTasks = append(actionTasks, *task)
	})

	return actionTasks
}

func (s *pipelineService) tryGetOrgName(p *spec.Pipeline) string {
	orgIDStr := p.MergeLabels()[apistructs.LabelOrgID]
	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		s.p.Log.Debugf("failed to parse orgID: %s, err: %v", orgIDStr, err)
		return ""
	}
	return s.cache.GetOrSetOrgName(orgID)
}

func initializePipelineGC(gc *basepb.PipelineGC) {
	if gc.DatabaseGC == nil {
		gc.DatabaseGC = &basepb.PipelineDatabaseGC{}
	}
	if gc.DatabaseGC.Analyzed == nil {
		gc.DatabaseGC.Analyzed = &basepb.PipelineDBGCItem{}
	}
	if gc.DatabaseGC.Finished == nil {
		gc.DatabaseGC.Finished = &basepb.PipelineDBGCItem{}
	}
	if gc.ResourceGC == nil {
		gc.ResourceGC = &basepb.PipelineResourceGC{}
	}
}
