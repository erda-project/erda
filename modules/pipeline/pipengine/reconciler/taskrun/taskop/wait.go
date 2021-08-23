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

package taskop

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/commonutil/costtimeutil"
	"github.com/erda-project/erda/modules/pipeline/conf"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/taskrun"
	"github.com/erda-project/erda/pkg/loop"
)

var err4EnableDeclineRatio = errors.New("enable decline ratio")

type wait taskrun.TaskRun

func NewWait(tr *taskrun.TaskRun) *wait {
	return (*wait)(tr)
}

func (w *wait) Op() taskrun.Op {
	return taskrun.Wait
}

func (w *wait) TaskRun() *taskrun.TaskRun {
	return (*taskrun.TaskRun)(w)
}

func (w *wait) Processing() (interface{}, error) {
	stopWaitCh := make(chan struct{})
	defer func() {
		stopWaitCh <- struct{}{}
	}()
	go func() {
		select {
		case <-w.Ctx.Done():
			w.StopWaitLoop = true
			return
		case <-w.PExitCh:
			logrus.Warnf("reconciler: pipeline exit, stop wait, pipelineID: %d, taskID: %d", w.P.ID, w.Task.ID)
			return
		case <-stopWaitCh:
			rlog.TDebugf(w.P.ID, w.Task.ID, "stop wait")
			close(stopWaitCh)
			return
		}
	}()

	var (
		data interface{}
	)

	err := loop.New(loop.WithDeclineRatio(1.5), loop.WithDeclineLimit(time.Second*10)).Do(func() (abort bool, err error) {
		if w.QuitWaitTimeout {
			return true, nil
		}

		statusDesc, err := w.Executor.Status(w.Ctx, w.Task)
		if err != nil {
			logrus.Errorf("[alert] reconciler: pipelineID: %d, task %q wait get status failed, err: %v",
				w.P.ID, w.Task.Name, err)
			return true, err
		}

		if statusDesc.Status == apistructs.PipelineStatusUnknown {
			logrus.Errorf("[alert] reconciler: pipelineID: %d, task %q wait get status %q, retry", w.P.ID, w.Task.Name, apistructs.PipelineStatusUnknown)
			return false, err4EnableDeclineRatio
		}

		if statusDesc.Status.IsEndStatus() {
			data = statusDesc
			return true, nil
		}

		return w.StopWaitLoop, err4EnableDeclineRatio
	})

	return data, err
}

func (w *wait) WhenDone(data interface{}) error {
	// go metrics.TaskEndEvent(*w.Task, w.P)
	if data == nil {
		return nil
	}
	endStatus := data.(apistructs.PipelineStatusDesc).Status
	if endStatus.IsFailedStatus() {
		if inspect, err := w.Executor.Inspect(w.Ctx, w.Task); err != nil {
			logrus.Errorf("failed to inspect task, pipelineID:%d, taskID: %d, err: %v", w.P.ID, w.Task.ID, err)
		} else {
			if inspect.Desc != "" {
				_ = w.TaskRun().UpdateTaskInspect(inspect.Desc)
			}
		}
	}
	w.Task.Status = endStatus
	w.Task.TimeEnd = time.Now()
	w.Task.CostTimeSec = costtimeutil.CalculateTaskCostTimeSec(w.Task)
	logrus.Infof("reconciler: pipelineID: %d, task %q end wait (%s -> %s, wait: %ds)",
		w.P.ID, w.Task.Name, apistructs.PipelineStatusRunning, data.(apistructs.PipelineStatusDesc).Status, w.Task.CostTimeSec)
	return nil
}

func (w *wait) WhenLogicError(err error) error {
	w.Task.Status = apistructs.PipelineStatusError
	return nil
}

func (w *wait) WhenTimeout() error {
	w.QuitQueueTimeout = true

	// 获取一次最新状态，轮训间隔期间可能任务已经是终态
	statusDesc, err := w.Executor.Status(w.Ctx, w.Task)
	if err == nil && statusDesc.Status.IsEndStatus() {
		w.FakeTimeout = true
		return w.WhenDone(statusDesc)
	}

	w.QuitWaitTimeout = true
	w.Task.Status = apistructs.PipelineStatusTimeout
	w.Task.TimeEnd = time.Now()
	w.Task.CostTimeSec = int64(w.Task.TimeEnd.Sub(w.Task.TimeBegin).Seconds())
	_, err = w.Executor.Cancel(w.Ctx, w.Task)
	return err
}

func (w *wait) TimeoutConfig() (<-chan struct{}, context.CancelFunc, time.Duration) {
	var timeoutCtx = context.Background()
	var cancel context.CancelFunc

	// -1: long run, no timeout limit
	// others: use task timeout
	taskTimeout := w.Task.Extra.Timeout
	if taskTimeout < -1 || taskTimeout == 0 { // < -1: invalid, 0: not set, use default
		taskTimeout = conf.TaskDefaultTimeout()
	}

	switch taskTimeout {
	case -1:
		// no limit
		return nil, nil, -1

	default:
		// set timeout
		var deadline time.Time
		if w.Task.TimeBegin.IsZero() {
			deadline = time.Now().Add(taskTimeout)
		} else {
			deadline = w.Task.TimeBegin.Add(taskTimeout)
		}
		// // 如果 deadline 在当前时刻之前，说明是 pipeline 平台挂了；
		// // 如果不对 deadline 做处理，则会立即收到 timeout 而导致无法查询到 task 真实状态；
		// // 需要将 deadline 设置为：当前时刻 + 一个足够查询一次 task 最新状态的时间
		// if deadline.Before(time.Now()) {
		// 	deadline = time.Now().Add(1 * time.Minute)
		// }
		timeoutCtx, cancel = context.WithDeadline(context.Background(), deadline)

		return timeoutCtx.Done(), cancel, taskTimeout
	}
}

func (w *wait) TuneTriggers() taskrun.TaskOpTuneTriggers {
	return taskrun.TaskOpTuneTriggers{
		BeforeProcessing: aoptypes.TuneTriggerTaskBeforeWait,
		AfterProcessing:  aoptypes.TuneTriggerTaskAfterWait,
	}
}
