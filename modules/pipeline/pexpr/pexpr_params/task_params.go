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

package pexpr_params

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/pvolumes"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/expression"
	"github.com/erda-project/erda/pkg/loop"
)

// GenerateParamsFromTask 生成 task 用于计算表达式的所有参数
// 包括：
// - 占位符参数
//   - configs.key
//   - outputs.preTaskName.key
//   - dirs.preTaskName.filepath
//   - params.key
// - 函数语法
//   - (echo hello world)
// - 内置变量
//   - pipeline_status
//   - task_status
func GenerateParamsFromTask(pipelineID uint64, taskID uint64, taskStatus apistructs.PipelineStatus) map[string]string {
	// get data from db
	var (
		p           *spec.Pipeline
		tasks       []*spec.PipelineTask
		currentTask *spec.PipelineTask
	)
	_ = loop.New(loop.WithDeclineRatio(2), loop.WithDeclineLimit(time.Second*10)).Do(func() (abort bool, err error) {
		pWithTasks, err := dbClient.GetPipelineWithTasks(pipelineID)
		if err != nil {
			logrus.Error(err)
			return false, err
		}
		p = pWithTasks.Pipeline
		tasks = pWithTasks.Tasks
		for _, task := range tasks {
			if task.ID == taskID {
				// because loop is before update current task db storage status
				task.Status = taskStatus
				currentTask = task
				break
			}
		}
		if currentTask == nil {
			return false, fmt.Errorf("currentTask must have value")
		}
		return true, nil
	})

	params := make(map[string]string)

	// outputs
	outputs := generateOutputs(tasks)
	for k, v := range outputs {
		params[k] = v
	}

	// configs
	configs := generateConfigs(p)
	for k, v := range configs {
		params[k] = v
	}

	// status
	params["pipeline_status"] = p.Status.String()
	for _, task := range tasks {
		if task.ID == taskID {
			params["task_status"] = task.Status.String()
		}
	}

	return params
}

// outputs: outputs.preTaskName.key
func generateOutputs(tasks []*spec.PipelineTask) map[string]string {
	makePhKeyFunc := func(taskName, metaKey string) string {
		return fmt.Sprintf(expression.Outputs+".%s.%s", taskName, metaKey)
	}
	outputs := make(map[string]string)
	for _, task := range tasks {
		for _, meta := range task.Result.Metadata {
			outputs[makePhKeyFunc(task.Name, meta.Name)] = meta.Value
		}
	}
	return outputs
}

// configs: configs.key
func generateConfigs(p *spec.Pipeline) map[string]string {
	makePhKeyFunc := func(key string) string { return fmt.Sprintf(expression.Configs+".%s", key) }
	configs := make(map[string]string)
	for k, v := range p.Snapshot.Secrets {
		configs[makePhKeyFunc(k)] = v
	}
	for k, v := range p.Snapshot.PlatformSecrets {
		configs[makePhKeyFunc(k)] = v
	}
	return configs
}

// workdirs:
//   - workdirs.preTaskName           只渲染到 workdir
//   - workdirs.preTaskName.filepath  workdir 后拼接用户指定的路径
func generateWorkdirs(tasks []spec.PipelineTask, currentTask spec.PipelineTask) map[string]string {
	makePhKeyFunc := func(taskName string) string { return fmt.Sprintf("workdirs.%s", taskName) }
	workdirs := make(map[string]string)
	for taskName, workdir := range pvolumes.GetAvailableTaskContainerWorkdirs(tasks, currentTask) {
		workdirs[makePhKeyFunc(taskName)] = workdir
	}
	return workdirs
}
