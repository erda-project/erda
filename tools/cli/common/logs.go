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

package common

import (
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"

	basepb "github.com/erda-project/erda-proto-go/core/pipeline/base/pb"
	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/tools/cli/command"
)

type PipelineTaskRef struct {
	ID         uint64
	PipelineID uint64
	StageID    uint64
	Name       string
	Type       string
	Status     string
	UUID       string
	StageIndex int
	TaskIndex  int
	TimeBegin  int64
}

type SelectTasksOptions struct {
	TaskID      uint64
	TaskName    string
	All         bool
	FailedOnly  bool
	RunningOnly bool
}

type TaskLogRequestOptions struct {
	PipelineID  uint64
	TaskID      uint64
	OrgName     string
	ClusterName string
	LogID       string
	Stream      string
	Tail        int
	Start       int64
	End         int64
	Count       int64
}

func FlattenPipelineTasks(pipeline pipelinepb.PipelineDetailDTO) []PipelineTaskRef {
	var tasks []PipelineTaskRef
	for stageIndex, stage := range pipeline.PipelineStages {
		if stage == nil {
			continue
		}
		for taskIndex, task := range stage.PipelineTasks {
			if task == nil || task.ID == 0 {
				continue
			}
			tasks = append(tasks, toPipelineTaskRef(stageIndex, taskIndex, task))
		}
	}
	return tasks
}

func SelectTasks(pipeline pipelinepb.PipelineDetailDTO, opts SelectTasksOptions) ([]PipelineTaskRef, error) {
	tasks := FlattenPipelineTasks(pipeline)

	switch {
	case opts.TaskID > 0:
		for _, task := range tasks {
			if task.ID == opts.TaskID {
				return []PipelineTaskRef{task}, nil
			}
		}
		return nil, errors.Errorf("task id %d not found in pipeline %d", opts.TaskID, pipeline.ID)
	case opts.TaskName != "":
		var matches []PipelineTaskRef
		for _, task := range tasks {
			if task.Name == opts.TaskName {
				matches = append(matches, task)
			}
		}
		if len(matches) == 0 {
			return nil, errors.Errorf("task name %q not found in pipeline %d", opts.TaskName, pipeline.ID)
		}
		if len(matches) > 1 {
			return nil, errors.Errorf("task name %q is ambiguous in pipeline %d", opts.TaskName, pipeline.ID)
		}
		return matches, nil
	case opts.FailedOnly:
		return filterTasks(tasks, func(task PipelineTaskRef) bool {
			return apistructs.PipelineStatus(task.Status).IsFailedStatus()
		}), nil
	case opts.RunningOnly:
		return filterTasks(tasks, func(task PipelineTaskRef) bool {
			status := apistructs.PipelineStatus(task.Status)
			return status.IsRunningStatus() || status.InQueue()
		}), nil
	case opts.All:
		return tasks, nil
	default:
		return tasks, nil
	}
}

func GetTaskLog(ctx *command.Context, opts TaskLogRequestOptions) (*apistructs.DashboardSpotLogData, error) {
	var resp apistructs.DashboardSpotLogResponse
	request := ctx.Get().
		Path(fmt.Sprintf("/api/cicd/%d/tasks/%d/logs", opts.PipelineID, opts.TaskID)).
		Header("org", opts.OrgName)

	for key, value := range buildTaskLogQueryParams(opts, time.Now()) {
		request = request.Param(key, value)
	}

	httpResp, err := request.Do().JSON(&resp)
	if err != nil {
		return nil, err
	}
	if !httpResp.IsOK() {
		return nil, errors.Errorf("status fail, status code: %d, err: %+v", httpResp.StatusCode(), resp.Error)
	}
	if !resp.Success {
		return nil, errors.Errorf("status fail: %+v", resp.Error)
	}

	return &resp.Data, nil
}

func buildTaskLogQueryParams(opts TaskLogRequestOptions, end time.Time) map[string]string {
	if opts.Tail <= 0 {
		opts.Tail = 200
	}
	count := -int64(opts.Tail)
	if opts.Count != 0 {
		count = opts.Count
	}
	endNano := end.UnixNano()
	if opts.End > 0 {
		endNano = opts.End
	}

	logID := opts.LogID
	if logID == "" && opts.TaskID > 0 {
		logID = fmt.Sprintf("pipeline-task-%d", opts.TaskID)
	}

	params := map[string]string{
		"count":  strconv.FormatInt(count, 10),
		"start":  strconv.FormatInt(opts.Start, 10),
		"end":    strconv.FormatInt(endNano, 10),
		"stream": opts.Stream,
		"taskID": strconv.FormatUint(opts.TaskID, 10),
		"id":     logID,
	}
	if opts.ClusterName != "" {
		params["clusterName"] = opts.ClusterName
	}
	return params
}

func filterTasks(tasks []PipelineTaskRef, predicate func(task PipelineTaskRef) bool) []PipelineTaskRef {
	var filtered []PipelineTaskRef
	for _, task := range tasks {
		if predicate(task) {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

func toPipelineTaskRef(stageIndex, taskIndex int, task *basepb.PipelineTaskDTO) PipelineTaskRef {
	ref := PipelineTaskRef{
		ID:         task.ID,
		PipelineID: task.PipelineID,
		StageID:    task.StageID,
		Name:       task.Name,
		Type:       task.Type,
		Status:     task.Status,
		StageIndex: stageIndex,
		TaskIndex:  taskIndex,
	}
	if task.Extra != nil {
		ref.UUID = task.Extra.UUID
	}
	if task.TimeBegin != nil {
		ref.TimeBegin = task.TimeBegin.AsTime().UnixNano()
	}
	return ref
}
