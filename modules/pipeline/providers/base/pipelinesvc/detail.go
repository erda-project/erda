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
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/costtimeutil"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
	"github.com/erda-project/erda/pkg/strutil"
)

func (s *PipelineSvc) Get(pipelineID uint64) (*spec.Pipeline, error) {
	p, err := s.dbClient.GetPipeline(pipelineID)
	if err != nil {
		return nil, apierrors.ErrGetPipeline.InternalError(err)
	}
	return &p, nil
}

func (s *PipelineSvc) SimplePipelineBaseDetail(pipelineID uint64) (*basepb.PipelineInstanceDetail, error) {
	base, find, err := s.dbClient.GetPipelineBase(pipelineID)
	if err != nil {
		return nil, err
	}
	if !find {
		return nil, fmt.Errorf("not find this pipeline id %v", pipelineID)
	}

	converted := s.convertPipelineBase(base)
	b, err := json.Marshal(converted)
	if err != nil {
		return nil, err
	}
	var detail basepb.PipelineInstanceDetail
	if err = json.Unmarshal(b, &detail); err != nil {
		return nil, err
	}

	return &detail, nil
}

func (s *PipelineSvc) Detail(pipelineID uint64) (*basepb.PipelineInstanceDetail, error) {
	p, err := s.dbClient.GetPipeline(pipelineID)

	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}

	p.CostTimeSec = costtimeutil.CalculatePipelineCostTimeSec(&p)

	// 创建时间特殊处理
	if len(p.Extra.CronExpr) > 0 {
		p.TimeCreated = p.Extra.CronTriggerTime
	}

	// 不展示 secret
	p.Snapshot.Secrets = nil

	stages, err := s.dbClient.ListPipelineStageByPipelineID(p.ID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}

	tasks, err := s.dbClient.ListPipelineTasksByPipelineID(p.ID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}

	var needApproval bool
	var stageDetailDTO []*basepb.PipelineStageDetail

	for _, stage := range stages {
		var taskDTOs []*basepb.PipelineTask
		for _, task := range tasks {
			if task.StageID != stage.ID {
				continue
			}
			if task.Type == "manual-review" {
				needApproval = true
			}
			task.CostTimeSec = costtimeutil.CalculateTaskCostTimeSec(&task)
			taskDTOs = append(taskDTOs, task.Convert2DTO())
		}
		stageDetailDTO = append(stageDetailDTO,
			&basepb.PipelineStageDetail{
				ID:            stage.ID,
				PipelineID:    stage.PipelineID,
				Name:          stage.Name,
				Status:        stage.Status.String(),
				CostTimeSec:   stage.CostTimeSec,
				TimeBegin:     timestamppb.New(stage.TimeBegin),
				TimeEnd:       timestamppb.New(stage.TimeEnd),
				TimeCreated:   timestamppb.New(stage.TimeCreated),
				TimeUpdated:   timestamppb.New(stage.TimeUpdated),
				PipelineTasks: taskDTOs,
			})
	}

	var pc *spec.PipelineCron
	// CronExpr 不为空，则 cron 必须存在
	if len(p.Extra.CronExpr) > 0 {
		if p.CronID == nil {
			return nil, apierrors.ErrGetPipelineDetail.MissingParameter("cronID")
		}
		c, err := s.dbClient.GetPipelineCron(*p.CronID)
		if err != nil {
			return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
		}
		pc = &c
	} else {
		// cron 按钮（开始、停止操作）目前挂在 pipeline 实例上，非周期创建的实例，也需要有 cron 按钮信息进行操作
		// 尝试根据 pipelineSource + pipelineYmlName 获取 cron
		pcs, _, err := s.dbClient.PagingPipelineCron(apistructs.PipelineCronPagingRequest{
			Sources:  []apistructs.PipelineSource{p.PipelineSource},
			YmlNames: []string{p.PipelineYmlName},
			PageSize: 1,
			PageNo:   1,
		})
		if err != nil {
			return nil, apierrors.ErrPagingPipelineCron.InternalError(err)
		}
		if len(pcs) > 0 {
			pc = &pcs[0]
		}
	}

	var detail basepb.PipelineInstanceDetail
	b, err := json.Marshal(s.ConvertPipeline(&p))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &detail); err != nil {
		return nil, err
	}
	detail.NeedApproval = needApproval

	// 插入 label
	pipelineLabels, _ := s.dbClient.ListLabelsByPipelineID(p.ID)
	labels := make(map[string]string, len(pipelineLabels))
	for _, v := range pipelineLabels {
		labels[v.Key] = v.Value
	}
	detail.Labels = labels
	detail.PipelineStages = stageDetailDTO
	detail.PipelineCron = pc.Convert2DTO()
	// 前端需要 cron 对象不为空
	if detail.PipelineCron == nil {
		detail.PipelineCron = &cronpb.Cron{}
	}

	buttons, err := s.setPipelineButtons(p, pc)
	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}

	detail.PipelineButton = buttons

	s.setPipelineTaskActionDetail(&detail)

	pipelineParams, err := getPipelineParams(p.PipelineYml, p.Snapshot.RunPipelineParams)
	if err != nil {
		return nil, err
	}
	detail.RunParams = pipelineParams

	// events
	detail.Events = s.getPipelineEvents(pipelineID)

	return &detail, nil
}

func getPipelineParams(pipelineYml string, runParams []apistructs.PipelineRunParamWithValue) ([]*basepb.PipelineParamWithValue, error) {

	pipeline, err := pipelineyml.New([]byte(pipelineYml))
	if err != nil {
		return nil, err
	}

	if pipeline == nil {
		return nil, errors.New("  getPipelineParams error: yml to pipeline error, pipeline is empty")
	}

	if pipeline.Spec() == nil {
		return nil, errors.New("  getPipelineParams error: pipeline spec is empty")
	}

	params := pipeline.Spec().Params
	if params == nil {
		return nil, nil
	}

	runParamsMap := make(map[string]*structpb.Value)
	if runParams != nil {
		for _, v := range runParams {
			vv, err := structpb.NewValue(v.Value)
			if err != nil {
				return nil, err
			}
			runParamsMap[v.Name] = vv
		}
	}

	var pipelineParamDTOs []*basepb.PipelineParamWithValue
	for _, param := range params {
		newDefault, err := structpb.NewValue(param.Default)
		if err != nil {
			return nil, err
		}
		pipelineParamDTOs = append(pipelineParamDTOs, &basepb.PipelineParamWithValue{
			Name:     param.Name,
			Desc:     param.Desc,
			Default:  newDefault,
			Required: param.Required,
			Type:     param.Type,
			Value:    runParamsMap[param.Name],
		})
	}
	return pipelineParamDTOs, nil
}

func (s *PipelineSvc) getPipelineEvents(pipelineID uint64) []*basepb.PipelineEvent {
	_, events, err := s.dbClient.GetPipelineEvents(pipelineID)
	if err != nil {
		logrus.Errorf("failed to get pipeline events, pipelineID: %d, err: %v", pipelineID, err)
		return nil
	}
	return events
}

// 给 pipelineTask 设置 action 的 logo 和 displayName 给前端展示
func (s *PipelineSvc) setPipelineTaskActionDetail(detail *basepb.PipelineInstanceDetail) {
	stageDetails := detail.PipelineStages

	var extensionSearchRequest = apistructs.ExtensionSearchRequest{}
	extensionSearchRequest.YamlFormat = true
	// 循环 stageDetails 数组，获取里面的 task 并设置到 Extensions 中
	loopStageDetails(stageDetails, func(task *basepb.PipelineTask) {
		extensionSearchRequest.Extensions = append(extensionSearchRequest.Extensions, task.Type)
	})
	if extensionSearchRequest.Extensions != nil {
		extensionSearchRequest.Extensions = strutil.DedupSlice(extensionSearchRequest.Extensions, true)
	}

	// 根据 Extensions 数组批量查询详情
	resultMap, err := s.bdl.SearchExtensions(extensionSearchRequest)
	if err != nil {
		logrus.Errorf("pipeline Detail to SearchExtensions error: %v", err)
		return
	}
	// 遍历 stageDetails 数组，根据 task 的 name 获取其 extension 详情
	var actionDetails = make(map[string]*basepb.PipelineTaskActionDetail)
	loopStageDetails(stageDetails, func(task *basepb.PipelineTask) {

		if task.Type == pipelineyml.Snippet {
			actionDetails[task.Type] = &basepb.PipelineTaskActionDetail{
				LogoUrl:     pipelineyml.SnippetLogo,
				DisplayName: pipelineyml.SnippetDisplayName,
				Description: pipelineyml.SnippetDesc,
			}
			return
		}

		version, ok := resultMap[task.Type]
		if !ok {
			return
		}

		specYmlStr, ok := version.Spec.(string)
		if !ok {
			return
		}

		var actionSpec apistructs.ActionSpec
		if err := yaml.Unmarshal([]byte(specYmlStr), &actionSpec); err != nil {
			return
		}

		actionDetails[task.Type] = &basepb.PipelineTaskActionDetail{
			LogoUrl:     actionSpec.LogoUrl,
			DisplayName: actionSpec.DisplayName,
			Description: actionSpec.Desc,
		}
	})
	detail.PipelineTaskActionDetails = actionDetails
}

// 遍历 stageDetails 数组，内部的每个 task 都执行一遍 doing 函数
func loopStageDetails(stageDetails []*basepb.PipelineStageDetail, doing func(dto *basepb.PipelineTask)) {
	if stageDetails != nil {
		for _, stage := range stageDetails {
			tasks := stage.PipelineTasks
			if tasks == nil {
				continue
			}

			for _, task := range tasks {
				doing(task)
			}
		}
	}
}

// Statistic pipeline 运行情况统计
func (s *PipelineSvc) Statistic(source, clusterName string) (*apistructs.PipelineStatisticResponseData, error) {
	return s.dbClient.PipelineStatistic(source, clusterName)
}

// 设置按钮状态
func (s *PipelineSvc) setPipelineButtons(p spec.Pipeline, pc *spec.PipelineCron) (button *basepb.PipelineButton, err error) {
	defer func() {
		err = errors.Wrap(err, "failed to set pipeline button")
	}()

	button = &basepb.PipelineButton{
		CanManualRun:   func() bool { _, can := s.canManualRun(p); return can }(),
		CanCancel:      canCancel(p),
		CanForceCancel: canForceCancel(p),
		CanRerun:       canRerun(p),
		CanRerunFailed: canRerunFailed(p),
		CanStartCron:   canStartCron(p, pc),
		CanStopCron:    canStopCron(p, pc),
		CanPause:       canPause(p),
		CanUnpause:     canUnpause(p),
		CanDelete:      func() bool { ok, _ := canDelete(p); return ok }(),
	}

	return
}

func (s *PipelineSvc) canManualRun(p spec.Pipeline) (reason string, can bool) {
	can = false

	if p.Status != apistructs.PipelineStatusAnalyzed {
		reason = fmt.Sprintf("pipeline already begin run")
		return
	}
	if p.Extra.ShowMessage != nil && p.Extra.ShowMessage.AbortRun {
		reason = "abort run, please check PreCheck result"
		return
	}
	if p.Type == apistructs.PipelineTypeRerunFailed && p.Extra.RerunFailedDetail != nil {
		rerunPipelineID := p.Extra.RerunFailedDetail.RerunPipelineID
		if rerunPipelineID > 0 {
			origin, err := s.dbClient.GetPipeline(rerunPipelineID)
			if err != nil {
				reason = fmt.Sprintf("failed to get origin pipeline when set canManualRun, rerunPipelineID: %d, err: %v", rerunPipelineID, err)
				return
			}
			if origin.Extra.CompleteReconcilerGC {
				reason = fmt.Sprintf("dependent rerun pipeline already been cleaned, rerunPipelineID: %d", rerunPipelineID)
				return
			}
		}
	}

	// default
	return "", true
}

func canCancel(p spec.Pipeline) bool {
	return p.Status.IsReconcilerRunningStatus()
}

// TODO 强制取消
func canForceCancel(p spec.Pipeline) bool {
	return false
}

// canRerun 重试全流程
func canRerun(p spec.Pipeline) bool {
	return p.Status.IsEndStatus()
}

// canRerunFailed 重试失败节点
func canRerunFailed(p spec.Pipeline) bool {
	// pipeline 状态为失败，且未被 gc 前，可以重试失败节点
	if p.Status.IsFailedStatus() && !p.Extra.CompleteReconcilerGC {
		return true
	}
	return false
}

// canStartCron p.cronID = pc.id
func canStartCron(p spec.Pipeline, pc *spec.PipelineCron) bool {
	return pc != nil && pc.Enable != nil && !*pc.Enable
}

// canStopCron p.cronID = pc.id
func canStopCron(p spec.Pipeline, pc *spec.PipelineCron) bool {
	return pc != nil && pc.Enable != nil && *pc.Enable
}

// canPause TODO 需要关心所有节点运行状态，如果所有节点都在运行中，则不能暂停
func canPause(p spec.Pipeline) bool {
	return false
}

// canUnpause TODO
func canUnpause(p spec.Pipeline) bool {
	return p.Status.CanUnpause()
}

func canDelete(p spec.Pipeline) (bool, string) {
	// status
	if !p.Status.CanDelete() {
		return false, fmt.Sprintf("invalid status: %s", p.Status)
	}
	// 终态后需要判断 complete gc
	if p.Status.IsEndStatus() {
		if !p.Extra.CompleteReconcilerGC {
			return false, fmt.Sprintf("waiting gc")
		}
	}
	return true, ""
}

func polishTask(p *spec.Pipeline, task *spec.PipelineTask, runningStageID uint64, dbClient *dbclient.Client) {
	changed := false
	defer func() {
		if changed {
			if err := dbClient.UpdatePipelineTaskStatus(task.ID, task.Status); err != nil {
				logrus.Errorf("[alert] failed to update pipeline task status when polishTask, pipelineID: %d, taskID: %d, err: %v",
					p.ID, task.ID, err)
			}
		}
	}()
	// 1. pipeline 为终态
	// 2. pipeline 非终态，只修改 stageID 小于 runningStageID 的任务
	if p.Status.IsEndStatus() || task.StageID < runningStageID {
		// task 仍在运行
		if task.Status == apistructs.PipelineStatusAnalyzed || task.Status.IsReconcilerRunningStatus() {
			// 整体成功，则 task 一定成功
			if p.Status.IsSuccessStatus() {
				task.Status = apistructs.PipelineStatusSuccess
				changed = true
				return
			}
			// 判断 task 状态
			if len(task.Result.Errors) > 0 {
				task.Status = apistructs.PipelineStatusFailed
				changed = true
				return
			}
			// 无法判断，判断为成功
			task.Status = apistructs.PipelineStatusSuccess
			changed = true
			return
		}
	}
}

// findRunningStageID 若 pipeline 为终态，返回 0
// R: Running A: Analyzed S: Success
// 1 R       1 R
// 2 S => 3  2 A => 1
// 3 S       3 A
func findRunningStageID(p spec.Pipeline, tasks []spec.PipelineTask) uint64 {
	if p.Status.IsEndStatus() {
		return 0
	}
	var runningStageID uint64 = 0
	for i := range tasks {
		task := tasks[i]
		if !task.Status.IsEndStatus() {
			if runningStageID == 0 {
				runningStageID = task.StageID
			}
			if task.StageID < runningStageID {
				runningStageID = task.StageID
			}
		} else { // task 终态
			if task.StageID > runningStageID {
				runningStageID = task.StageID
			}
		}
	}
	return runningStageID
}
