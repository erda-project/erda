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
	"context"
	"runtime/debug"
	"sync"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/statusutil"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor/types"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/taskrun"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

// reconcile do pipeline reconcile.
func (r *Reconciler) reconcile(ctx context.Context, pipelineID uint64) error {
	// judge if dlock lost
	if ctx.Err() != nil {
		rlog.PWarnf(pipelineID, "no need reconcile, dlock already lost, err: %v", ctx.Err())
		return nil
	}
	// init caches and get stages
	defer clearPipelineContextCaches(pipelineID)

	// get latest pipeline before reconcile
	p, err := r.dbClient.GetPipeline(pipelineID)
	if err != nil {
		return err
	}
	if p.Status.IsEndStatus() {
		rlog.PWarnf(pipelineID, "pipeline is already end status (%s), invoke ctx.done directly", p.Status)
		ctx.Done()
		return nil
	}

	tasks, err := r.YmlTaskMergeDBTasks(&p)
	if err != nil {
		return err
	}

	// delay gc if have
	r.delayGC(p.Extra.Namespace, p.ID)

	// calculate pipeline status by tasks
	calcPStatus := statusutil.CalculatePipelineStatusV2(tasks)
	logrus.Infof("reconciler: pipelineID: %d, pipeline is not completed, continue reconcile, currentStatus: %s",
		p.ID, p.Status)

	schedulableTasks, err := r.getSchedulableTasks(&p, tasks)
	if err != nil {
		return rlog.PErrorAndReturn(p.ID, err)
	}

	var wg sync.WaitGroup
	for i := range schedulableTasks {
		wg.Add(1)
		go func(i int) {
			var err error
			defer func() {
				if r := recover(); r != nil {
					debug.PrintStack()
					err = errors.Errorf("%v", r)
				}
				if err != nil {
					logrus.Errorf("[alert] reconciler: pipelineID: %d, task %q reconcile occurred an error: %v", p.ID, schedulableTasks[i].Name, err)
				}
				r.processingTasks.Delete(buildTaskDagName(p.ID, schedulableTasks[i].Name))
				err = r.reconcile(ctx, pipelineID)
				wg.Done()
			}()

			var task *spec.PipelineTask
			task, err = r.saveTask(schedulableTasks[i], &p)
			if err != nil {
				logrus.Errorf("[alert] reconciler: pipelineID: %d, task %q failed to save task: %v", p.ID, schedulableTasks[i].Name, err)
				return
			}

			if task.IsSnippet {
				task, err = r.reconcileSnippetTask(task, &p)
				return
			}

			executor, err := actionexecutor.GetManager().Get(types.Name(task.Extra.ExecutorName))
			if err != nil {
				return
			}
			tr := taskrun.New(ctx, task,
				ctx.Value(ctxKeyPipelineExitCh).(chan struct{}), ctx.Value(ctxKeyPipelineExitChCancelFunc).(context.CancelFunc),
				r.TaskThrottler, executor, &p, r.bdl, r.dbClient, r.js,
				r.actionAgentSvc, r.extMarketSvc)

			// tear down task
			defer func() {
				if tr.Task.Status.IsEndStatus() {
					// 同步 teardown
					tr.Teardown()
				}
			}()

			// 从 executor 获取最新任务状态，防止重复创建、启动的情况发生
			latestStatusFromExecutor, err := tr.Executor.Status(tr.Ctx, tr.Task)
			if err == nil && tr.Task.Status != latestStatusFromExecutor.Status {
				if latestStatusFromExecutor.Status.IsAbnormalFailedStatus() {
					logrus.Errorf("[alert] reconciler: pipelineID: %d, task %q, not correct task status from executor: %s -> %s (abnormal), continue reconcile task",
						p.ID, tr.Task.Name, tr.Task.Status, latestStatusFromExecutor.Status)
				} else {
					logrus.Errorf("[alert] reconciler: pipelineID: %d, task %q, correct task status from executor: %s -> %s",
						p.ID, tr.Task.Name, tr.Task.Status, latestStatusFromExecutor.Status)
					tr.Task.Status = latestStatusFromExecutor.Status
					tr.Update()
				}
			}

			// 之前的节点有失败的, 然后 action 中没有 if 表达式，直接更新状态为失败
			if calcPStatus == apistructs.PipelineStatusFailed && tr.Task.Extra.Action.If == "" {
				tr.Task.Status = apistructs.PipelineStatusNoNeedBySystem
				tr.Task.Extra.AllowFailure = true
				tr.Update()
				return
			}

			err = reconcileTask(tr)
			return
		}(i)
	}
	wg.Wait()

	return nil
}
