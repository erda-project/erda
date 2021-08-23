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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/events"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (s *PipelineSvc) DealPipelineCallbackOfAction(data []byte) (err error) {
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "failed to deal with pipeline action callback")
		}
	}()

	// 回调数据格式校验
	var cb apistructs.ActionCallback
	if err = json.Unmarshal(data, &cb); err != nil {
		return err
	}
	if cb.PipelineTaskID <= 0 {
		return errors.Errorf("invalid pipelineTaskID [%d]", cb.PipelineTaskID)
	}

	task, err := s.dbClient.GetPipelineTask(cb.PipelineTaskID)
	if err != nil {
		return err
	}
	p, err := s.dbClient.GetPipeline(task.PipelineID)
	if err != nil {
		return err
	}

	if task.PipelineID != p.ID {
		return apierrors.ErrCallback.InvalidParameter(
			fmt.Sprintf("task not belong to pipeline, taskID: %d, pipelineID: %d", task.ID, p.ID))
	}

	// 更新 task.result
	if err = s.appendPipelineTaskResult(&p, &task, cb); err != nil {
		return err
	}

	// 处理特殊回调逻辑
	// 1. runtimeID
	if err = s.doCallbackOfRuntimeID(&p, &task, cb); err != nil {
		return err
	}
	// 2. flink/spark jar resource
	if err = s.doCallbackOfJarResource(&p, &task, cb); err != nil {
		return err
	}

	return nil
}

// appendPipelineTaskResult 追加 result
func (s *PipelineSvc) appendPipelineTaskResult(p *spec.Pipeline, task *spec.PipelineTask, cb apistructs.ActionCallback) error {
	if len(cb.Metadata) == 0 && len(cb.Errors) == 0 && cb.MachineStat == nil {
		return nil
	}
	// metadata
	task.Result.Metadata = append(task.Result.Metadata, cb.Metadata...)
	// TODO action agent should add err start time and end time
	newTaskErrors := make([]*apistructs.PipelineTaskErrResponse, 0)
	for _, e := range cb.Errors {
		newTaskErrors = append(newTaskErrors, &apistructs.PipelineTaskErrResponse{
			Msg: e.Msg,
		})
	}
	task.Result.Errors = task.Result.AppendError(newTaskErrors...)
	// machine stat
	if cb.MachineStat != nil {
		task.Result.MachineStat = cb.MachineStat
	}

	if err := s.dbClient.UpdatePipelineTaskResult(task.ID, task.Result); err != nil {
		return err
	}

	// emit event when meta updated
	events.EmitTaskEvent(task, p)

	return nil
}

// doCallbackOfRuntimeID 发送 websocket 消息，及时更新页面 link
func (s *PipelineSvc) doCallbackOfRuntimeID(p *spec.Pipeline, task *spec.PipelineTask, cb apistructs.ActionCallback) error {
	for _, meta := range cb.Metadata {
		if meta.Type == apistructs.ActionCallbackTypeLink &&
			meta.Name == apistructs.ActionCallbackRuntimeID {
			events.EmitTaskRuntimeEvent(task, p)
			break
		}
	}
	return nil
}

// doCallbackOfJarResource 获取 flink/spark 任务需要的 jar resource
func (s *PipelineSvc) doCallbackOfJarResource(p *spec.Pipeline, task *spec.PipelineTask, cb apistructs.ActionCallback) error {
	for _, meta := range cb.Metadata {
		if meta.Name != "bigdataJarResource" {
			continue
		}
		// 寻找需要这个 task 生成的 jar resource 的 flink/spark task
		flinkSparkTasks, err := s.findFlinkSparkTasks(p, task.Name)
		if err != nil {
			return err
		}
		for _, fst := range flinkSparkTasks {
			fst.Extra.FlinkSparkConf.JarResource = meta.Value
			if err = s.dbClient.UpdatePipelineTask(fst.ID, &fst); err != nil {
				return err
			}
		}
	}
	return nil
}

// findFlinkSparkTasks 寻找 depend 为指定值的 task
func (s *PipelineSvc) findFlinkSparkTasks(p *spec.Pipeline, depend string) ([]spec.PipelineTask, error) {
	tasks, err := s.dbClient.ListPipelineTasksByPipelineID(p.ID)
	if err != nil {
		return nil, err
	}
	var result []spec.PipelineTask
	for i := range tasks {
		task := tasks[i]
		if isFlinkSparkAction(task.Type) && task.Extra.FlinkSparkConf.Depend == depend && len(task.Extra.FlinkSparkConf.JarResource) == 0 {
			result = append(result, task)
		}
	}
	return result, nil
}

func isFlinkSparkAction(action string) bool {
	return action == "flink" || action == "spark"
}
