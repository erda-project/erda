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
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/parser/pipelineyml"
)

// makeNormalPipelineTask 生成普通流水线任务
func (s *PipelineSvc) makeNormalPipelineTask(p *spec.Pipeline, ps *spec.PipelineStage, action *pipelineyml.Action) (*spec.PipelineTask, error) {
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
	executorKind, executorName, err := s.judgeTaskExecutor(action)
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
func (s *PipelineSvc) judgeTaskExecutor(action *pipelineyml.Action) (spec.PipelineTaskExecutorKind, string, error) {
	if action.Type == apistructs.ActionTypeAPITest {
		return spec.PipelineTaskExecutorKindAPITest, spec.PipelineTaskExecutorNameAPITestDefault, nil
	}
	return spec.PipelineTaskExecutorKindScheduler, spec.PipelineTaskExecutorNameSchedulerDefault, nil
}
