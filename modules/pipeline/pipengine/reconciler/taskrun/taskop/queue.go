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

package taskop

import (
	"context"
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

type queue taskrun.TaskRun

func NewQueue(tr *taskrun.TaskRun) *queue {
	return (*queue)(tr)
}

func (q *queue) Op() taskrun.Op {
	return taskrun.Queue
}

func (q *queue) TaskRun() *taskrun.TaskRun {
	return (*taskrun.TaskRun)(q)
}

func (q *queue) Processing() (interface{}, error) {
	stopQueueCh := make(chan struct{})
	defer func() {
		stopQueueCh <- struct{}{}
	}()
	go func() {
		select {
		case <-q.Ctx.Done():
			q.StopQueueLoop = true
			return
		case <-q.PExitCh:
			logrus.Warnf("reconciler: pipeline exit, stop queue, pipelineID: %d, taskID: %d", q.P.ID, q.Task.ID)
			return
		case <-stopQueueCh:
			rlog.TDebugf(q.P.ID, q.Task.ID, "stop queue")
			close(stopQueueCh)
			return
		}
	}()

	var lastMessage string

	beginQueue := q.Task.Extra.TimeBeginQueue
	if beginQueue.IsZero() {
		beginQueue = time.Now()
	}
	alerted := false

	err := loop.New(loop.WithDeclineRatio(1.5), loop.WithDeclineLimit(time.Second*10)).Do(func() (abort bool, err error) {
		// 排队超时钉钉告警
		if !alerted && time.Now().Sub(beginQueue) > conf.TaskQueueAlertTime() {
			logrus.Errorf("[pipeline] task queue exceed %v, beginTime: %s, now: %s, cluster: %s, pipelineID: %d, taskID: %d, taskName: %s, reason: %s",
				conf.TaskQueueAlertTime(), beginQueue.Format(time.RFC3339), time.Now().Format(time.RFC3339),
				q.P.ClusterName, q.P.ID, q.Task.ID, q.Task.Name, lastMessage)
			alerted = true
		}

		statusDesc, err := q.Executor.Status(q.Ctx, q.Task)
		if err != nil {
			return true, err
		}

		newMsg := statusDesc.Desc
		if newMsg != "" && newMsg != lastMessage {
			_ = q.TaskRun().AppendLastMsg(newMsg)
			lastMessage = newMsg
		}

		if !statusDesc.Status.IsEndStatus() && statusDesc.Status != apistructs.PipelineStatusRunning {
			if inspect, _ := q.Executor.Inspect(q.Ctx, q.Task); inspect.Desc != "" {
				_ = q.TaskRun().UpdateTaskInspect(inspect.Desc)
			}
		}

		if statusDesc.Status == apistructs.PipelineStatusUnknown {
			logrus.Errorf("[alert] reconciler: pipelineID: %d, task %q queue get status %q, retry", q.P.ID, q.Task.Name, apistructs.PipelineStatusUnknown)
			return false, err4EnableDeclineRatio
		}

		if statusDesc.Status == apistructs.PipelineStatusRunning || statusDesc.Status.IsEndStatus() {
			return true, nil
		}

		return q.StopQueueLoop, err4EnableDeclineRatio
	})

	return nil, err
}

func (q *queue) WhenDone(data interface{}) error {
	q.Task.Status = apistructs.PipelineStatusRunning
	q.Task.Extra.TimeEndQueue = time.Now()
	q.Task.QueueTimeSec = costtimeutil.CalculateTaskQueueTimeSec(q.Task)
	q.Task.TimeBegin = time.Now()
	logrus.Infof("reconciler: pipelineID: %d, task %q end queue (%s -> %s, queue: %ds)",
		q.P.ID, q.Task.Name, apistructs.PipelineStatusQueue, apistructs.PipelineStatusRunning, q.Task.QueueTimeSec)
	return nil
}

func (q *queue) WhenLogicError(err error) error {
	q.Task.Status = apistructs.PipelineStatusError
	return nil
}

func (q *queue) WhenTimeout() error {
	return nil
}

func (q *queue) TimeoutConfig() (<-chan struct{}, context.CancelFunc, time.Duration) {
	return nil, nil, -1
}

func (q *queue) TuneTriggers() taskrun.TaskOpTuneTriggers {
	return taskrun.TaskOpTuneTriggers{
		BeforeProcessing: aoptypes.TuneTriggerTaskBeforeQueue,
		AfterProcessing:  aoptypes.TuneTriggerTaskAfterQueue,
	}
}
