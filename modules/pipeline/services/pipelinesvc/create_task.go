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

package pipelinesvc

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/services/extmarketsvc"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/numeral"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

// makeNormalPipelineTask 生成普通流水线任务
func (s *PipelineSvc) makeNormalPipelineTask(p *spec.Pipeline, ps *spec.PipelineStage, action *pipelineyml.Action, passedData passedDataWhenCreate) (*spec.PipelineTask, error) {
	var actionJobDefine = passedData.getActionJobDefine(extmarketsvc.MakeActionTypeVersion(action))
	var actionJobSpec = passedData.getActionJobSpecs(extmarketsvc.MakeActionTypeVersion(action))

	task := &spec.PipelineTask{}
	task.PipelineID = p.ID
	task.StageID = ps.ID
	task.Name = action.Alias.String()
	// task.OpType
	task.Type = action.Type.String()
	task.Extra.Namespace = p.Extra.Namespace
	task.Extra.ClusterName = p.ClusterName
	task.Extra.AllowFailure = false
	task.Extra.Pause = false
	task.Extra.Timeout = time.Duration(action.Timeout * int64(time.Second))
	if action.Timeout < 0 {
		task.Extra.Timeout = time.Duration(action.Timeout)
	}
	task.Extra.StageOrder = ps.Extra.StageOrder
	// task.Extra.Envs
	// task.Extra.Labels
	// task.Extra.Image

	// set executor
	executorKind, executorName, err := s.judgeTaskExecutor(action, actionJobSpec)
	if err != nil {
		return nil, apierrors.ErrCreatePipelineTask.InvalidParameter(err)
	}
	task.ExecutorKind = executorKind
	task.Extra.ExecutorName = executorName

	// default task resource limit
	task.Extra.RuntimeResource = spec.RuntimeResource{
		CPU:    conf.TaskDefaultCPU(),
		Memory: conf.TaskDefaultMEM(),
		Disk:   0,
	}
	// task.Extra.SelfInputs
	// task.Extra.SelfOutputs
	if isFlinkSparkAction(action.Type.String()) {
		task.Extra.FlinkSparkConf = spec.FlinkSparkConf{
			Depend:    getString(action.Params["depends"]),
			MainClass: getString(action.Params["main_class"]),
			MainArgs:  []string{getString(action.Params["main_args"])},
		}
	}

	task.Extra.Action = *action

	// runAfter
	for _, need := range action.Needs {
		task.Extra.RunAfter = append(task.Extra.RunAfter, need.String())
	}

	task.Status = apistructs.PipelineStatusAnalyzed
	if task.Extra.Pause {
		task.Status = apistructs.PipelineStatusPaused
	}
	task.CostTimeSec = -1
	task.QueueTimeSec = -1

	// 给 task 设置上 snippet action 定制的 env
	if action.SnippetConfig != nil && action.SnippetConfig.Labels != nil {
		actionEnv := action.SnippetConfig.Labels[apistructs.LabelActionEnv]
		var actionEnvLabels = map[string]string{}
		err := json.Unmarshal([]byte(actionEnv), &actionEnvLabels)
		if err == nil {
			for key, v := range actionEnvLabels {
				if task.Extra.PrivateEnvs == nil {
					task.Extra.PrivateEnvs = map[string]string{}
				}
				task.Extra.PrivateEnvs[key] = v
			}
		} else {
			logrus.Errorf("error load action %v snippetConfig", action)
		}
	}

	// applied resources
	task.Extra.AppliedResources = calculateNormalTaskResources(action, actionJobDefine)

	return task, nil
}

// makeSnippetPipelineTask 生成嵌套流水线任务
// action: 从 yaml 解析出来的 action 信息
// p: 当前层的 pipeline，已先于 task 创建好
// stage: stage 信息，已先于 task 创建好
func (s *PipelineSvc) makeSnippetPipelineTask(p *spec.Pipeline, stage *spec.PipelineStage, action *pipelineyml.Action) (*spec.PipelineTask, error) {
	var task spec.PipelineTask
	task.PipelineID = p.ID
	task.StageID = stage.ID
	task.Name = action.Alias.String()
	task.Type = apistructs.ActionTypeSnippet
	task.ExecutorKind = spec.PipelineTaskExecutorKindScheduler
	task.Status = apistructs.PipelineStatusAnalyzed

	// extra
	extra, err := s.genSnippetTaskExtra(p, action)
	if err != nil {
		return nil, fmt.Errorf("failed to generate snippet task extra, pipelineID: %d, actionName: %s, err: %v", p.ID, action.Alias.String(), err)
	}
	task.Extra = extra

	// snippet
	task.IsSnippet = true

	task.CostTimeSec = -1
	task.QueueTimeSec = -1

	// ext resources set outside after created

	return &task, nil
}

func (s *PipelineSvc) genSnippetTaskExtra(p *spec.Pipeline, action *pipelineyml.Action) (spec.PipelineTaskExtra, error) {
	var ex spec.PipelineTaskExtra
	ex.Namespace = p.Extra.Namespace
	ex.ExecutorName = spec.PipelineTaskExecutorNameEmpty
	ex.ClusterName = p.ClusterName
	ex.AllowFailure = false
	ex.Pause = false
	ex.Timeout = s.calculateTaskTimeoutDuration(action)
	ex.PrivateEnvs = nil
	ex.PublicEnvs = nil
	ex.Labels = nil
	ex.RuntimeResource = spec.GenDefaultTaskResource()
	ex.RunAfter = s.calculateTaskRunAfter(action)
	ex.Action = *action
	return ex, nil
}

func (s *PipelineSvc) calculateTaskTimeoutDuration(action *pipelineyml.Action) time.Duration {
	if action.Timeout == pipelineyml.TimeoutDuration4Forever {
		return pipelineyml.TimeoutDuration4Forever
	}
	return time.Duration(action.Timeout * int64(time.Second))
}

func (s *PipelineSvc) calculateTaskRunAfter(action *pipelineyml.Action) []string {
	var runAfters []string
	for _, need := range action.Needs {
		runAfters = append(runAfters, need.String())
	}
	return runAfters
}

// judgeTaskExecutor judge task executor by action info
func (s *PipelineSvc) judgeTaskExecutor(action *pipelineyml.Action, actionSpec *apistructs.ActionSpec) (spec.PipelineTaskExecutorKind, spec.PipelineTaskExecutorName, error) {
	if actionSpec == nil ||
		actionSpec.Executor == nil ||
		len(actionSpec.Executor.Kind) <= 0 ||
		len(actionSpec.Executor.Name) <= 0 ||
		!spec.PipelineTaskExecutorKind(actionSpec.Executor.Kind).Check() ||
		!spec.PipelineTaskExecutorName(actionSpec.Executor.Name).Check() {
		return spec.PipelineTaskExecutorKindScheduler, spec.PipelineTaskExecutorNameSchedulerDefault, nil
	}

	return spec.PipelineTaskExecutorKind(actionSpec.Executor.Kind), spec.PipelineTaskExecutorName(actionSpec.Executor.Name), nil
}

func calculateNormalTaskResources(action *pipelineyml.Action, actionDefine *diceyml.Job) apistructs.PipelineAppliedResources {
	defaultRes := apistructs.PipelineAppliedResource{CPU: conf.TaskDefaultCPU(), MemoryMB: conf.TaskDefaultMEM()}
	return apistructs.PipelineAppliedResources{
		Limits:   calculateNormalTaskLimitResource(action, actionDefine, defaultRes),
		Requests: calculateNormalTaskRequestResource(action, actionDefine, defaultRes),
	}
}

func calculateNormalTaskLimitResource(action *pipelineyml.Action, actionDefine *diceyml.Job, defaultRes apistructs.PipelineAppliedResource) apistructs.PipelineAppliedResource {
	// calculate
	maxCPU := numeral.MaxFloat64([]float64{
		actionDefine.Resources.MaxCPU, actionDefine.Resources.CPU,
		action.Resources.MaxCPU, action.Resources.CPU,
	})
	maxMemoryMB := numeral.MaxFloat64([]float64{
		float64(actionDefine.Resources.MaxMem), float64(actionDefine.Resources.Mem),
		float64(action.Resources.Mem),
	})

	// use default if is empty
	if maxCPU == 0 {
		maxCPU = defaultRes.CPU
	}
	if maxMemoryMB == 0 {
		maxMemoryMB = defaultRes.MemoryMB
	}

	return apistructs.PipelineAppliedResource{
		CPU:      maxCPU,
		MemoryMB: maxMemoryMB,
	}
}

func calculateNormalTaskRequestResource(action *pipelineyml.Action, actionDefine *diceyml.Job, defaultRes apistructs.PipelineAppliedResource) apistructs.PipelineAppliedResource {
	// assign from actionDefine
	requestCPU := numeral.MinFloat64([]float64{actionDefine.Resources.MaxCPU, actionDefine.Resources.CPU}, true)
	requestMemoryMB := numeral.MinFloat64([]float64{float64(actionDefine.Resources.MaxMem), float64(actionDefine.Resources.Mem)}, true)

	// user explicit declaration has the highest priority, overwrite value from actionDefine
	if c := numeral.MinFloat64([]float64{action.Resources.MaxCPU, action.Resources.CPU}, true); c > 0 {
		requestCPU = c
	}
	if m := action.Resources.Mem; m > 0 {
		requestMemoryMB = float64(m)
	}

	// use default if is empty
	if requestCPU == 0 {
		requestCPU = defaultRes.CPU
	}
	if requestMemoryMB == 0 {
		requestMemoryMB = defaultRes.MemoryMB
	}

	return apistructs.PipelineAppliedResource{
		CPU:      requestCPU,
		MemoryMB: requestMemoryMB,
	}
}
