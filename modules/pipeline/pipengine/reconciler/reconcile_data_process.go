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

package reconciler

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/events"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// parse out tasks according to the yml structure, and then query the created tasks from the database,
// and replace the tasks that already exist in the database with yml tasks
func (r *Reconciler) YmlTaskMergeDBTasks(pipeline *spec.Pipeline) ([]*spec.PipelineTask, error) {
	// get pipeline tasks from db
	tasks, err := r.dbClient.ListPipelineTasksByPipelineID(pipeline.ID)
	if err != nil {
		return nil, err
	}

	// get or set stages from caches
	stages, err := getOrSetStagesFromContext(r.dbClient, pipeline.ID)
	if err != nil {
		return nil, err
	}

	// get or set pipelineYml from caches
	pipelineYml, err := getOrSetPipelineYmlFromContext(r.dbClient, pipeline.ID)
	if err != nil {
		return nil, err
	}

	passedDataWhenCreate, err := getOrSetPassedDataWhenCreateFromContext(r.bdl, pipelineYml, pipeline.ID)
	if err != nil {
		return nil, err
	}

	tasks, err = r.pipelineSvcFunc.MergePipelineYmlTasks(pipelineYml, tasks, pipeline, stages, passedDataWhenCreate)
	if err != nil {
		return nil, err
	}

	var newTasks []*spec.PipelineTask
	for index := range tasks {
		task := tasks[index]
		newTasks = append(newTasks, &task)
	}

	return newTasks, err
}

func (r *Reconciler) saveTask(task *spec.PipelineTask, pipeline *spec.Pipeline) (*spec.PipelineTask, error) {
	if task.ID > 0 {
		return task, nil
	}

	lastSuccessTaskMap, err := getOrSetPipelineRerunSuccessTasksFromContext(r.dbClient, pipeline.ID)
	if err != nil {
		return nil, err
	}

	stages, err := getOrSetStagesFromContext(r.dbClient, pipeline.ID)
	if err != nil {
		return nil, err
	}

	var pt *spec.PipelineTask
	lastSuccessTask, ok := lastSuccessTaskMap[task.Name]
	if ok {
		pt = lastSuccessTask
		pt.ID = 0
		pt.PipelineID = pipeline.ID
		pt.StageID = stages[task.Extra.StageOrder].ID
	} else {
		pt = task
	}

	// save action
	if err := r.dbClient.CreatePipelineTask(pt); err != nil {
		logrus.Errorf("[alert] failed to create pipeline task when create pipeline graph: %v", err)
		return nil, err
	}

	return pt, nil
}

func (r *Reconciler) createSnippetPipeline(task *spec.PipelineTask, p *spec.Pipeline) (snippetPipeline *spec.Pipeline, resultTask *spec.PipelineTask, err error) {
	var failedError error
	defer func() {
		if failedError != nil {
			err = failedError
			task.Result.Errors = append(task.Result.Errors, &apistructs.PipelineTaskErrResponse{
				Msg: err.Error(),
			})
			task.Status = apistructs.PipelineStatusFailed
			if updateErr := r.dbClient.UpdatePipelineTask(task.ID, task); updateErr != nil {
				err = updateErr
				return
			}
			snippetPipeline = nil
			resultTask = nil
		}
	}()
	var taskSnippetConfig = apistructs.SnippetConfig{
		Source: task.Extra.Action.SnippetConfig.Source,
		Name:   task.Extra.Action.SnippetConfig.Name,
		Labels: task.Extra.Action.SnippetConfig.Labels,
	}
	sourceSnippetConfigs := []apistructs.SnippetConfig{taskSnippetConfig}
	sourceSnippetConfigYamls, err := r.pipelineSvcFunc.HandleQueryPipelineYamlBySnippetConfigs(sourceSnippetConfigs)
	if err != nil {
		failedError = err
		return nil, nil, failedError
	}
	if len(sourceSnippetConfigYamls) <= 0 {
		return nil, nil, fmt.Errorf("not find snippet %v yml", taskSnippetConfig.ToString())
	}

	snippetPipeline, err = r.pipelineSvcFunc.MakeSnippetPipeline4Create(p, task, sourceSnippetConfigYamls[taskSnippetConfig.ToString()])
	if err != nil {
		return nil, nil, err
	}
	if err := r.pipelineSvcFunc.CreatePipelineGraph(snippetPipeline); err != nil {
		return nil, nil, err
	}

	task.SnippetPipelineID = &snippetPipeline.ID
	task.Extra.AppliedResources = snippetPipeline.Snapshot.AppliedResources
	if err := r.dbClient.UpdatePipelineTask(task.ID, task); err != nil {
		return nil, nil, err
	}
	return snippetPipeline, task, nil
}

func (r *Reconciler) reconcileSnippetTask(task *spec.PipelineTask, p *spec.Pipeline) (resultTask *spec.PipelineTask, err error) {
	defer func() {
		if err != nil && errors.Is(dbclient.NotFoundBaseError, err) {
			if err := r.dbClient.UpdatePipelineTaskStatus(task.ID, apistructs.PipelineStatusError); err != nil {
				rlog.TErrorf(p.ID, task.ID, "failed to update snippet task status to Error when not found base, err: %v", err)
			}
		}
	}()
	var snippetPipeline *spec.Pipeline
	if task.SnippetPipelineID != nil && *task.SnippetPipelineID > 0 {
		snippetPipelineValue, err := r.dbClient.GetPipeline(task.SnippetPipelineID)
		if err != nil {
			return nil, err
		}
		snippetPipeline = &snippetPipelineValue
	} else {
		snippetPipeline, task, err = r.createSnippetPipeline(task, p)
		if err != nil {
			return nil, err
		}
	}

	if snippetPipeline == nil {
		task.Status = apistructs.PipelineStatusAnalyzeFailed
		task.Result.Errors = append(task.Result.Errors, &apistructs.PipelineTaskErrResponse{
			Msg: "not find task bind pipeline",
		})
		if updateErr := r.dbClient.UpdatePipelineTask(task.ID, task); updateErr != nil {
			err = updateErr
			return nil, updateErr
		}
		return nil, fmt.Errorf("not find task bind pipeline")
	}

	sp := snippetPipeline
	// make context for snippet
	snippetCtx := makeContextForPipelineReconcile(sp.ID)
	// snippet pipeline first run
	if sp.Status == apistructs.PipelineStatusAnalyzed {
		// copy pipeline level run info from root pipeline
		if err = r.copyParentPipelineRunInfo(sp); err != nil {
			return nil, err
		}
		// set begin time
		now := time.Now()
		sp.TimeBegin = &now
		if err = r.dbClient.UpdatePipelineBase(sp.ID, &sp.PipelineBase); err != nil {
			return nil, err
		}

		var snippetPipelineTasks []*spec.PipelineTask
		snippetPipelineTasks, err = r.YmlTaskMergeDBTasks(sp)
		if err != nil {
			return nil, err
		}
		snippetDetail := apistructs.PipelineTaskSnippetDetail{
			DirectSnippetTasksNum:    len(snippetPipelineTasks),
			RecursiveSnippetTasksNum: -1,
		}
		if err := r.dbClient.UpdatePipelineTaskSnippetDetail(task.ID, snippetDetail); err != nil {
			return nil, err
		}

		if err := r.dbClient.UpdatePipelineTaskStatus(task.ID, apistructs.PipelineStatusRunning); err != nil {
			return nil, err
		}
	}

	if err := r.updateStatusBeforeReconcile(*sp); err != nil {
		rlog.PErrorf(p.ID, "Failed to update pipeline status before reconcile, err: %v", err)
		return nil, err
	}
	err = r.reconcile(snippetCtx, sp.ID)
	defer func() {
		r.teardownCurrentReconcile(snippetCtx, sp.ID)
		if err := r.updateStatusAfterReconcile(snippetCtx, sp.ID); err != nil {
			logrus.Errorf("snippet pipeline: %d, failed to update status after reconcile, err: %v", sp.ID, err)
		}
	}()
	if err != nil {
		return nil, err
	}
	// 查询最新 task
	latestTask, err := r.dbClient.GetPipelineTask(task.ID)
	if err != nil {
		return nil, err
	}
	*task = *(&latestTask)
	return task, nil
}

// updatePipeline update db, publish websocket event
func (r *Reconciler) updatePipelineStatus(p *spec.Pipeline) error {
	// db
	if err := r.dbClient.UpdatePipelineBaseStatus(p.ID, p.Status); err != nil {
		return err
	}

	// event
	events.EmitPipelineInstanceEvent(p, p.GetRunUserID())

	return nil
}

func buildTaskDagName(pipelineID uint64, taskName string) string {
	return strconv.FormatUint(pipelineID, 10) + "_" + taskName
}
