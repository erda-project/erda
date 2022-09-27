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
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	cronpb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	common "github.com/erda-project/erda-proto-go/core/pipeline/pb"
	"github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/pipeline/commonutil/costtimeutil"
	"github.com/erda-project/erda/internal/tools/pipeline/dbclient"
	"github.com/erda-project/erda/internal/tools/pipeline/pkg/taskresult"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/cron/crontypes"
	"github.com/erda-project/erda/internal/tools/pipeline/services/apierrors"
	"github.com/erda-project/erda/internal/tools/pipeline/spec"
	"github.com/erda-project/erda/pkg/i18n"
	"github.com/erda-project/erda/pkg/metadata"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

func (s *pipelineService) Get(pipelineID uint64) (*spec.Pipeline, error) {
	p, err := s.dbClient.GetPipeline(pipelineID)
	if err != nil {
		return nil, apierrors.ErrGetPipeline.InternalError(err)
	}
	return &p, nil
}

func (s *pipelineService) PipelineDetail(ctx context.Context, req *pb.PipelineDetailRequest) (*pb.PipelineDetailResponse, error) {
	var detailDTO *pb.PipelineDetailDTO
	var err error

	if req.SimplePipelineBaseResult {
		detailDTO, err = s.SimplePipelineBaseDetail(req.PipelineID)
	} else {
		detailDTO, err = s.Detail(req.PipelineID)
	}

	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}

	return &pb.PipelineDetailResponse{
		Data: detailDTO,
	}, nil
}

func (s *pipelineService) SimplePipelineBaseDetail(pipelineID uint64) (*pb.PipelineDetailDTO, error) {
	base, find, err := s.dbClient.GetPipelineBase(pipelineID)
	if err != nil {
		return nil, err
	}
	if !find {
		return nil, fmt.Errorf("not find this pipeline id %v", pipelineID)
	}

	var detail pb.PipelineDetailDTO
	baseDetail := s.ConvertPipelineBase(base)
	detail.ID = baseDetail.ID
	detail.Source = baseDetail.Source
	detail.YmlName = baseDetail.YmlName
	detail.Namespace = baseDetail.Namespace
	detail.ClusterName = baseDetail.ClusterName
	detail.Status = baseDetail.Status
	detail.Type = baseDetail.Type
	detail.TriggerMode = baseDetail.TriggerMode
	detail.CronID = baseDetail.CronID
	detail.Labels = baseDetail.Labels
	detail.YmlSource = baseDetail.YmlSource
	detail.Extra = baseDetail.Extra
	detail.OrgID = baseDetail.OrgID
	detail.OrgName = baseDetail.OrgName
	detail.ProjectID = baseDetail.ProjectID
	detail.ProjectName = baseDetail.ProjectName
	detail.ApplicationID = baseDetail.ApplicationID
	detail.ApplicationName = baseDetail.ApplicationName
	detail.Branch = baseDetail.Branch
	detail.Commit = baseDetail.Commit
	detail.CommitDetail = baseDetail.CommitDetail
	detail.Progress = baseDetail.Progress
	detail.CostTimeSec = baseDetail.CostTimeSec
	detail.Progress = baseDetail.Progress
	detail.TimeCreated = baseDetail.TimeCreated
	detail.TimeEnd = baseDetail.TimeEnd
	detail.TimeUpdated = baseDetail.TimeUpdated
	detail.TimeBegin = baseDetail.TimeBegin

	return &detail, nil
}

func (s *pipelineService) Detail(pipelineID uint64) (*pb.PipelineDetailDTO, error) {
	p, err := s.dbClient.GetPipeline(pipelineID)

	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}

	p.CostTimeSec = costtimeutil.CalculatePipelineCostTimeSec(&p)

	// 创建时间特殊处理
	if len(p.Extra.CronExpr) > 0 {
		p.TimeCreated = p.Extra.CronTriggerTime
	}

	stages, err := s.dbClient.ListPipelineStageByPipelineID(p.ID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}

	tasks, err := s.dbClient.ListPipelineTasksByPipelineID(p.ID)
	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}

	// init yml
	pipelineYml, err := pipelineyml.New(
		[]byte(p.PipelineYml),
	)
	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}
	// merge yml task and db task
	tasks, err = s.MergePipelineYmlTasks(pipelineYml, tasks, &p, stages, nil)
	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}

	actionMap, err := s.getPipelineActionMap(tasks)
	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}

	var needApproval bool
	var stageDetailDTO []*basepb.PipelineStageDetailDTO
	actionTasks := s.GetYmlActionTasks(pipelineYml, &p, stages, nil)

	for _, stage := range stages {
		var taskDTOs []*basepb.PipelineTaskDTO
		for _, task := range tasks {
			if task.StageID != stage.ID {
				continue
			}
			if task.Type == "manual-review" {
				needApproval = true
			}
			task.CostTimeSec = costtimeutil.CalculateTaskCostTimeSec(&task)
			if task.Result == nil {
				task.Result = &taskresult.Result{}
				task.Result.Metadata = make([]metadata.MetadataField, 0)
			}
			// add task events to result metadata if task status isn`t success and events it`s failed
			if !task.Status.IsSuccessStatus() && task.Inspect.Events != "" && !isEventsLatestNormal(task.Inspect.Events) {
				task.Result.Metadata = append(task.Result.Metadata, metadata.MetadataField{
					Name:  "task-events",
					Value: task.Inspect.Events,
				})
			}
			// set analyzed yaml task to NoNeedBySystem to simulate db task's behaviour
			if p.Status.IsStopByUser() && task.Status == apistructs.PipelineStatusAnalyzed {
				task.Status = apistructs.PipelineStatusNoNeedBySystem
			}
			taskDTO := task.Convert2PB()
			var actionSpec apistructs.ActionSpec
			var ymlTask spec.PipelineTask
			if action, ok := actionMap[task.Type]; ok {
				if specYmlStr, ok := action.Spec.(string); ok {
					if err := yaml.Unmarshal([]byte(specYmlStr), &actionSpec); err != nil {
						logrus.Errorf("unmarshal action spec error: %v, continue merge task param", err)
					}
					taskDTO.Extra.Action = actionSpec.Convert2PBDetail()
				}
			}
			for _, actionTask := range actionTasks {
				if actionTask.Name == task.Name {
					ymlTask = actionTask
					break
				}
			}
			taskDTO.Extra.Params = task.MergeTaskParamDetailToDisplay(actionSpec, ymlTask, p.Snapshot)
			taskDTOs = append(taskDTOs, taskDTO)
		}
		stageDetail := *stage.Convert2DTO()
		stageDetailPB := basepb.PipelineStageDetailDTO{
			ID:            stageDetail.ID,
			PipelineID:    stageDetail.PipelineID,
			Name:          stageDetail.Name,
			Status:        stageDetail.Status.String(),
			CostTimeSec:   stageDetail.CostTimeSec,
			TimeBegin:     timestamppb.New(stageDetail.TimeBegin),
			TimeEnd:       timestamppb.New(stageDetail.TimeEnd),
			TimeUpdated:   timestamppb.New(stageDetail.TimeUpdated),
			TimeCreated:   timestamppb.New(stageDetail.TimeCreated),
			PipelineTasks: taskDTOs,
		}
		stageDetailDTO = append(stageDetailDTO, &stageDetailPB)
	}

	// no need to display secret
	p.Snapshot.Secrets = nil

	var pc *common.Cron
	// CronExpr 不为空，则 cron 必须存在
	if len(p.Extra.CronExpr) > 0 {
		if p.CronID == nil {
			return nil, apierrors.ErrGetPipelineDetail.MissingParameter("cronID")
		}

		result, err := s.cronSvc.CronGet(context.Background(), &cronpb.CronGetRequest{
			CronID: *p.CronID,
		})
		if err != nil {
			return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
		}
		if result.Data == nil {
			return nil, apierrors.ErrNotFoundPipelineCron.InternalError(crontypes.ErrCronNotFound)
		}
		pc = result.Data
	} else {
		// cron 按钮（开始、停止操作）目前挂在 pipeline 实例上，非周期创建的实例，也需要有 cron 按钮信息进行操作
		// 尝试根据 pipelineSource + pipelineYmlName 获取 cron
		resp, err := s.cronSvc.CronPaging(context.Background(), &cronpb.CronPagingRequest{
			Sources:  []string{p.PipelineSource.String()},
			YmlNames: []string{p.PipelineYmlName},
			PageSize: 1,
			PageNo:   1,
		})
		if err != nil {
			return nil, apierrors.ErrPagingPipelineCron.InternalError(err)
		}
		if len(resp.Data) > 0 {
			pc = resp.Data[0]
		}
	}

	var detail pb.PipelineDetailDTO
	detail.NeedApproval = needApproval
	baseDetail := s.ConvertPipeline(&p)
	detail.ID = baseDetail.ID
	detail.Source = baseDetail.Source
	detail.YmlName = baseDetail.YmlName
	detail.Namespace = baseDetail.Namespace
	detail.ClusterName = baseDetail.ClusterName
	detail.Status = baseDetail.Status
	detail.Type = baseDetail.Type
	detail.TriggerMode = baseDetail.TriggerMode
	detail.CronID = baseDetail.CronID
	detail.Labels = baseDetail.Labels
	detail.YmlSource = baseDetail.YmlSource
	detail.YmlContent = baseDetail.YmlContent
	detail.Extra = baseDetail.Extra
	detail.OrgID = baseDetail.OrgID
	detail.OrgName = baseDetail.OrgName
	detail.ProjectID = baseDetail.ProjectID
	detail.ProjectName = baseDetail.ProjectName
	detail.ApplicationID = baseDetail.ApplicationID
	detail.ApplicationName = baseDetail.ApplicationName
	detail.Branch = baseDetail.Branch
	detail.Commit = baseDetail.Commit
	detail.CommitDetail = baseDetail.CommitDetail
	detail.Progress = baseDetail.Progress
	detail.CostTimeSec = baseDetail.CostTimeSec
	detail.Progress = baseDetail.Progress
	detail.TimeCreated = baseDetail.TimeCreated
	detail.TimeEnd = baseDetail.TimeEnd
	detail.TimeUpdated = baseDetail.TimeUpdated
	detail.TimeBegin = baseDetail.TimeBegin

	// 插入 label
	pipelineLabels, _ := s.dbClient.ListLabelsByPipelineID(p.ID)
	labels := make(map[string]string, len(pipelineLabels))
	for _, v := range pipelineLabels {
		labels[v.Key] = v.Value
	}
	detail.Labels = labels
	detail.PipelineStages = stageDetailDTO
	detail.PipelineCron = pc
	// 前端需要 cron 对象不为空
	if detail.PipelineCron == nil {
		detail.PipelineCron = &common.Cron{}
	}

	buttons, err := s.setPipelineButtons(p, pc)
	if err != nil {
		return nil, apierrors.ErrGetPipelineDetail.InternalError(err)
	}

	detail.PipelineButton = &buttons

	s.setPipelineTaskActionDetail(&detail, actionMap)

	pipelineParams, err := getPipelineParams(p.PipelineYml, p.Snapshot.RunPipelineParams)
	if err != nil {
		return nil, err
	}
	detail.RunParams = pipelineParams

	// events
	detail.Events = s.getPipelineEvents(pipelineID)

	return &detail, nil
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

	// 1. pipeline is final state
	// 2. The pipeline is not final, only the tasks whose stageID is less than runningStageID are modified
	if p.Status.IsEndStatus() || task.StageID < runningStageID {
		// task is still running
		if task.Status == apistructs.PipelineStatusAnalyzed || task.Status.IsReconcilerRunningStatus() {
			// the overall success, then the task must be successful
			if p.Status.IsSuccessStatus() {
				task.Status = apistructs.PipelineStatusSuccess
				changed = true
				return
			}
			// judge task status
			if len(task.Inspect.Errors) > 0 {
				task.Status = apistructs.PipelineStatusFailed
				changed = true
				return
			}
			// unable to judge, judged as successful
			task.Status = apistructs.PipelineStatusSuccess
			changed = true
			return
		}
	}
}

// findRunningStageID return 0 if pipeline is final
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
		} else { // task is end status
			if task.StageID > runningStageID {
				runningStageID = task.StageID
			}
		}
	}
	return runningStageID
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

	runParamsMap := make(map[string]interface{})
	if runParams != nil {
		for _, v := range runParams {
			runParamsMap[v.Name] = v.Value
		}
	}

	var pipelineParamDTOs []*basepb.PipelineParamWithValue
	for _, param := range params {
		defaultVal, err := structpb.NewValue(param.Default)
		if err != nil {
			return nil, err
		}
		val, err := structpb.NewValue(runParamsMap[param.Name])
		if err != nil {
			return nil, err
		}
		pipelineParamDTOs = append(pipelineParamDTOs, &basepb.PipelineParamWithValue{
			Name:     param.Name,
			Desc:     param.Desc,
			Default:  defaultVal,
			Required: param.Required,
			Type:     param.Type,
			Value:    val,
		})
	}
	return pipelineParamDTOs, nil
}

func (s *pipelineService) getPipelineEvents(pipelineID uint64) []*basepb.PipelineEvent {
	_, events, err := s.dbClient.GetPipelineEvents(pipelineID)
	if err != nil {
		logrus.Errorf("failed to get pipeline events, pipelineID: %d, err: %v", pipelineID, err)
		return nil
	}
	return events
}

// setPipelineTaskActionDetail set the action's logo and displayName for the pipelineTask to display to the front end
func (s *pipelineService) setPipelineTaskActionDetail(detail *pb.PipelineDetailDTO, actionMap map[string]apistructs.ExtensionVersion) {
	stageDetails := detail.PipelineStages

	// 遍历 stageDetails 数组，根据 task 的 name 获取其 extension 详情
	var actionDetails = make(map[string]*basepb.PipelineTaskActionDetail)
	loopStageDetails(stageDetails, func(task *basepb.PipelineTaskDTO) {

		if task.Type == pipelineyml.Snippet {
			actionDetails[task.Type] = &basepb.PipelineTaskActionDetail{
				LogoUrl:     pipelineyml.SnippetLogo,
				DisplayName: pipelineyml.SnippetDisplayName,
				Description: pipelineyml.SnippetDesc,
			}
			return
		}

		version, ok := actionMap[task.Type]
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
			DisplayName: actionSpec.GetLocaleDisplayName(i18n.GetGoroutineBindLang()),
			Description: actionSpec.GetLocaleDesc(i18n.GetGoroutineBindLang()),
		}
	})
	detail.PipelineTaskActionDetails = actionDetails
}

// loopStageDetails traverse the stageDetails array, each internal task executes the doing function once
func loopStageDetails(stageDetails []*basepb.PipelineStageDetailDTO, doing func(dto *basepb.PipelineTaskDTO)) {
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

func (s *pipelineService) getPipelineActionMap(tasks []spec.PipelineTask) (map[string]apistructs.ExtensionVersion, error) {
	var extensionSearchRequest = apistructs.ExtensionSearchRequest{}
	extensionSearchRequest.YamlFormat = true
	for _, task := range tasks {
		if task.Type == apistructs.ActionTypeSnippet {
			continue
		}
		extensionSearchRequest.Extensions = append(extensionSearchRequest.Extensions, task.Type)
	}
	resultMap, err := s.bdl.SearchExtensions(extensionSearchRequest)
	if err != nil {
		logrus.Errorf("pipeline Detail to SearchExtensions error: %v", err)
		return map[string]apistructs.ExtensionVersion{}, err
	}
	return resultMap, nil
}

func canCancel(p spec.Pipeline) bool {
	return p.Status.IsReconcilerRunningStatus()
}

// TODO Force cancellation
func canForceCancel(p spec.Pipeline) bool {
	return false
}

// canRerun retry the whole process
func canRerun(p spec.Pipeline) bool {
	return p.Status.IsEndStatus()
}

// canRerunFailed retry failed node
func canRerunFailed(p spec.Pipeline) bool {
	// pipeline status is failed, and the failed node can be retried before it is not gc
	if p.Status.IsFailedStatus() && !p.Extra.CompleteReconcilerGC {
		return true
	}
	return false
}

// canStartCron p.cronID = pc.id
func canStartCron(p spec.Pipeline, pc *common.Cron) bool {
	return pc != nil && pc.Enable != nil && !pc.Enable.Value
}

// canStopCron p.cronID = pc.id
func canStopCron(p spec.Pipeline, pc *common.Cron) bool {
	return pc != nil && pc.Enable != nil && pc.Enable.Value
}

// canPause TODO need to care about the running status of all nodes. If all nodes are running, you cannot pause
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
	// after the final state, it is necessary to judge the complete gc
	if p.Status.IsEndStatus() {
		if !p.Extra.CompleteReconcilerGC {
			return false, fmt.Sprintf("waiting gc")
		}
	}
	return true, ""
}

// setPipelineButtons set button state
func (s *pipelineService) setPipelineButtons(p spec.Pipeline, pc *common.Cron) (button basepb.PipelineButton, err error) {
	defer func() {
		err = errors.Wrap(err, "failed to set pipeline button")
	}()

	button = basepb.PipelineButton{
		CanManualRun:   func() bool { _, can := s.run.CanManualRun(context.Background(), &p); return can }(),
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

//isEventsContainWarn return k8s events is contain warn
//Events:
// Type    Reason     Age   From               Message
// ----    ------     ----  ----               -------
// Normal  Scheduled  7s    default-scheduler  Successfully assigned pipeline-4152/pipeline-4152.pipeline-task-8296-tgxd7 to node-010000006200
// Normal  Pulled     6s    kubelet            Container image "registry.erda.cloud/erda-actions/action-agent:1.2-20210804-75232495" already present on machine
func isEventsContainWarn(events string) bool {
	eventLst := strings.Split(events, "\n")
	if len(eventLst) <= 3 {
		return false
	}
	for _, ev := range eventLst {
		if strings.Contains(ev, corev1.EventTypeWarning) {
			return true
		}
	}
	return false
}

// isEventsLatestNormal judge k8s events latest event is normal
func isEventsLatestNormal(events string) bool {
	eventLst := strings.Split(events, "\n")
	if len(eventLst) <= 3 {
		return true
	}
	return strings.Contains(eventLst[len(eventLst)-2], corev1.EventTypeNormal)
}
