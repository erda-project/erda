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

package reconciler

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/commonutil/statusutil"
	"github.com/erda-project/erda/modules/pipeline/events"
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

	// get latest pipeline before reconcile
	pipelineWithTasks, err := r.dbClient.GetPipelineWithTasks(pipelineID)
	if err != nil {
		rlog.PErrorf(pipelineID, "cannot reconcile, failed to get pipeline with tasks, err: %v", err)
	}
	p := pipelineWithTasks.Pipeline
	tasks := pipelineWithTasks.Tasks

	if p.Status.IsEndStatus() {
		if p.Extra.CompleteReconcilerTeardown {
			return nil
		}
		// teardown pipeline
		r.teardownPipeline(ctx, pipelineWithTasks)
	}
	defer func() {
		// already end status, no need reconcile, return
		if p.Status.IsEndStatus() {
			if p.Extra.CompleteReconcilerTeardown {
				return
			}
			r.teardownPipeline(ctx, pipelineWithTasks)
		}
	}()

	// delay gc if have
	r.delayGC(p.Extra.Namespace, p.ID)

	// calculate pipeline status by tasks
	calcPStatus := statusutil.CalculatePipelineStatusV2(tasks)
	// 所有状态都为终止状态的时候
	if statusutil.CalculatePipelineTaskAllDone(tasks) {
		if p.Status != calcPStatus && !p.Status.IsEndStatus() {
			oldStatus := p.Status
			p.Status = calcPStatus
			if err := r.updatePipelineStatus(p); err != nil {
				return err
			}
			logrus.Infof("reconciler: pipelineID: %d, update pipeline status (%s -> %s)", p.ID, oldStatus, calcPStatus)
		}
	} else {
		// 直接更新为 running 状态
		if p.Status == apistructs.PipelineStatusAnalyzed || p.Status == apistructs.PipelineStatusQueue {
			oldStatus := p.Status
			p.Status = apistructs.PipelineStatusRunning
			if err := r.updatePipelineStatus(p); err != nil {
				return err
			}
			logrus.Infof("reconciler: pipelineID: %d, update pipeline status (%s -> %s)", p.ID, oldStatus, apistructs.PipelineStatusRunning)
			// go metrics.PipelineGaugeProcessingAdd(*p, 1)
		}
	}

	if p.Status.IsEndStatus() {
		return nil
	}

	logrus.Infof("reconciler: pipelineID: %d, pipeline is not completed, continue reconcile, currentStatus: %s",
		p.ID, p.Status)

	schedulableTasks, err := r.getSchedulableTasks(p, tasks)
	if err != nil {
		return rlog.PErrorAndReturn(p.ID, err)
	}

	var wg sync.WaitGroup
	for i := range schedulableTasks {
		task := schedulableTasks[i]

		wg.Add(1)

		go func() {
			var err error
			defer func() {
				if r := recover(); r != nil {
					debug.PrintStack()
					err = errors.Errorf("%v", r)
				}
				if err != nil {
					logrus.Errorf("[alert] reconciler: pipelineID: %d, task %q reconcile occurred an error: %v", p.ID, task.Name, err)
				}
				r.processingTasks.Delete(task.ID)
				err = r.reconcile(ctx, pipelineID)
				wg.Done()
			}()

			// 嵌套流水线
			if task.IsSnippet {
				snippetPipelineWithTasks, sErr := r.dbClient.GetPipelineWithTasks(*task.SnippetPipelineID)
				if sErr != nil {
					if strings.Contains(sErr.Error(), "not found") {
						err = fmt.Errorf("%s, no need retry(not found)", sErr)
						task.Status = apistructs.PipelineStatusAnalyzeFailed
						if updateErr := r.dbClient.UpdatePipelineTask(task.ID, task); updateErr != nil {
							err = updateErr
							return
						}
						return
					}
					err = sErr
					return
				}
				sp := snippetPipelineWithTasks.Pipeline
				// 第一次执行，赋予初始值
				if sp.Status == apistructs.PipelineStatusAnalyzed {
					// copy pipeline level run info from root pipeline
					if err = r.copyParentPipelineRunInfo(sp); err != nil {
						return
					}
					// set begin time
					now := time.Now()
					sp.TimeBegin = &now
					if err = r.dbClient.UpdatePipelineBase(snippetPipelineWithTasks.Pipeline.ID, &sp.PipelineBase); err != nil {
						return
					}
				}
				// 更新 task 状态为 running
				task.Status = apistructs.PipelineStatusRunning
				if err = r.dbClient.UpdatePipelineTaskStatus(task.ID, task.Status); err != nil {
					return
				}
				// 更新 task snippet detail
				snippetDetail := apistructs.PipelineTaskSnippetDetail{
					DirectSnippetTasksNum:    len(snippetPipelineWithTasks.Tasks),
					RecursiveSnippetTasksNum: -1,
				}
				if err := r.dbClient.UpdatePipelineTaskSnippetDetail(task.ID, snippetDetail); err != nil {
					return
				}
				// make context for snippet
				snippetCtx := makeContextForPipelineReconcile(sp.ID)
				if err = r.reconcile(snippetCtx, sp.ID); err != nil {
					return
				}
				// 查询最新 task
				latestTask, err := r.dbClient.GetPipelineTask(task.ID)
				if err != nil {
					return
				}
				*task = *(&latestTask)
				return
			}

			executor, err := actionexecutor.GetManager().Get(types.Name(task.Extra.ExecutorName))
			if err != nil {
				return
			}

			tr := taskrun.New(ctx, task,
				ctx.Value(ctxKeyPipelineExitCh).(chan struct{}), ctx.Value(ctxKeyPipelineExitChCancelFunc).(context.CancelFunc),
				r.TaskThrottler, executor, p, r.bdl, r.dbClient, r.js,
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
		}()
	}
	wg.Wait()

	return nil
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
