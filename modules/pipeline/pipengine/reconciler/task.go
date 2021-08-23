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
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/taskrun"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/taskrun/taskop"
	"github.com/erda-project/erda/modules/pipeline/pkg/errorsx"
	"github.com/erda-project/erda/pkg/loop"
)

var (
	defaultRetryDeclineRatio    = 2
	defaultRetryIntervalSec     = 30
	defaultRetryDeclineLimitSec = 600
)

func reconcileTask(tr *taskrun.TaskRun) error {
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "start reconcile task")
	defer rlog.TDebugf(tr.P.ID, tr.Task.ID, "end reconcile task")
	// // do metric
	// go metrics.TaskGaugeProcessingAdd(*tr.Task, 1)
	// defer func() {
	// 	go metrics.TaskGaugeProcessingAdd(*tr.Task, -1)
	// }()
	// do aop
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "start do task aop")
	if err := aop.Handle(aop.NewContextForTask(*tr.Task, *tr.P, aoptypes.TuneTriggerTaskBeforeExec)); err != nil {
		rlog.TErrorf(tr.P.ID, tr.Task.ID, "failed to handle aop, type: %s, err: %v", aoptypes.TuneTriggerTaskBeforeExec, err)
	}
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "end do task aop")

	// 系统异常时重试3次
	//const abnormalErrMaxRetryTimes = 3

	var platformErrRetryTimes int
	//var sessionNotFoundTimes int

	for {
		// stop reconciler task if pipeline stopping reconcile
		if tr.PExit {
			rlog.TWarnf(tr.P.ID, tr.Task.ID, "pipeline stopping reconcile, so stop reconcile task")
			return nil
		}

		var taskOp taskrun.TaskOp

		// create -> start -> queue -> wait -> end
		switch tr.Task.Status {

		// prepare
		case apistructs.PipelineStatusAnalyzed:
			taskOp = taskop.NewPrepare(tr)

		// create
		case apistructs.PipelineStatusBorn:
			taskOp = taskop.NewCreate(tr)

		// start
		case apistructs.PipelineStatusCreated:
			taskOp = taskop.NewStart(tr)

		// queue
		case apistructs.PipelineStatusQueue:
			taskOp = taskop.NewQueue(tr)

		// wait
		case apistructs.PipelineStatusRunning:
			taskOp = taskop.NewWait(tr)
		}

		if taskOp != nil {
			// Do 没有 err 才是正常的，即使失败也是 status=Failed，没有 error
			abnormalErr := tr.Do(taskOp)
			if abnormalErr != nil {
				if errorsx.IsContainUserError(abnormalErr) {
					rlog.TErrorf(tr.P.ID, tr.Task.ID, "failed to handle taskOp: %s, user abnormalErr: %v, don't need retry", taskOp.Op(), abnormalErr)
					return abnormalErr
				}
				// don't contain user error mean err is platform error, should retry always
				rlog.TErrorf(tr.P.ID, tr.Task.ID, "failed to handle taskOp: %s, abnormalErr: %v, continue retry, retry times: %d", taskOp.Op(), abnormalErr, platformErrRetryTimes)
				resetTaskForAbnormalRetry(tr, platformErrRetryTimes)
				platformErrRetryTimes++
				continue
			}
			// 没有异常，执行后续逻辑
		}

		if tr.Task.Status.IsEndStatus() {
			return nil
		}
	}
}

// resetTaskForAbnormalRetry 重置 task 后进行异常重试
func resetTaskForAbnormalRetry(tr *taskrun.TaskRun, abnormalErrRetryTimes int) {
	defaultInterval := time.Second * 30
	// 计算思考时间
	interval := defaultInterval
	if tr.Task.Extra.LoopOptions != nil && tr.Task.Extra.LoopOptions.CalculatedLoop != nil {
		strategy := tr.Task.Extra.LoopOptions.CalculatedLoop.Strategy
		interval = loop.New(
			loop.WithInterval(time.Second*time.Duration(strategy.IntervalSec)),
			loop.WithDeclineRatio(strategy.DeclineRatio),
			loop.WithDeclineLimit(time.Second*time.Duration(strategy.DeclineLimitSec)),
		).CalculateInterval(uint64(abnormalErrRetryTimes))
	} else {
		interval = loop.New(
			loop.WithInterval(time.Second*time.Duration(defaultRetryIntervalSec)),
			loop.WithDeclineRatio(float64(defaultRetryDeclineRatio)),
			loop.WithDeclineLimit(time.Second*time.Duration(defaultRetryDeclineLimitSec)),
		).CalculateInterval(uint64(abnormalErrRetryTimes))
	}
	if interval < defaultInterval {
		interval = defaultInterval
	}
	// 等待思考时间
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "sleep %s before retry abnormal retry", interval.String())
	time.Sleep(interval)

	tr.Update()
}
