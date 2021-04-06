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
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/taskrun"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/taskrun/taskop"
	"github.com/erda-project/erda/pkg/loop"
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
	const abnormalErrMaxRetryTimes = 3

	var platformErrRetryTimes int

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
				rlog.TWarnf(tr.P.ID, tr.Task.ID, "failed to handle taskOp: %s, abnormalErr: %v, continue retry", taskOp.Op(), abnormalErr)
				// 小于异常重试次数，继续重试
				if platformErrRetryTimes < abnormalErrMaxRetryTimes {
					resetTaskForAbnormalRetry(tr, platformErrRetryTimes)
					platformErrRetryTimes++
					continue
				}
				// 大于重试次数仍有异常，返回异常
				rlog.TErrorf(tr.P.ID, tr.Task.ID, "failed to handle taskOp: %s, abnormalErr: %v, reach max retry times, return abnormalErr", taskOp.Op(), abnormalErr)
				return abnormalErr
			}
			// 没有异常，执行后续逻辑
		}

		// 非终态，继续推进
		if !tr.Task.Status.IsEndStatus() {
			continue
		}

		// 循环
		if err := handleTaskLoop(tr); err != nil {
			// 作为异常重试
			if platformErrRetryTimes < abnormalErrMaxRetryTimes {
				platformErrRetryTimes++
				continue
			}
			return err
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
	}
	if interval < defaultInterval {
		interval = defaultInterval
	}
	// 等待思考时间
	rlog.TDebugf(tr.P.ID, tr.Task.ID, "sleep %s before retry abnormal retry", interval.String())
	time.Sleep(interval)

	// 更新状态
	tr.Task.Status = apistructs.PipelineStatusAnalyzed
	tr.Update()
}
