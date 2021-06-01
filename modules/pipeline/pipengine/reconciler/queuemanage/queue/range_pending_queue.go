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

package queue

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/aop"
	"github.com/erda-project/erda/modules/pipeline/aop/aoptypes"
	"github.com/erda-project/erda/modules/pipeline/events"
	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/priorityqueue"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/spec"
	"github.com/erda-project/erda/pkg/loop"
)

const (
	EventComponentQueueManager = "QueueManager"
)

// Queue event reason list
const (
	PendingQueueValidate = "PendingQueueValidate"
	FailedQueue          = "FailedQueue"
	SuccessQueue         = "SuccessQueue"
)

func (q *defaultQueue) RangePendingQueue() {
	q.setIsRangingPendingQueueFlag()
	defer q.unsetIsRangingPendingQueueFlag()
	usage := q.Usage()
	usageByte, _ := json.Marshal(&usage)
	logrus.Debugf("queueManager: queueID: %s, queueName: %s, usage: %s", q.ID(), q.pq.Name, string(usageByte))
	// fast reRange
	defer func() {
		if q.needReRangePendingQueue() {
			logrus.Debugf("queueManager: queueID: %s, queueName: %s, unset need reRange pending queue flag", q.ID(), q.pq.Name)
			// unset flag for next use
			q.unsetNeedReRangePendingQueueFlag()
		}
	}()
	q.eq.PendingQueue().Range(func(item priorityqueue.Item) (stopRange bool) {
		// fast reRange
		defer func() {
			if q.needReRangePendingQueue() {
				// stop current range
				stopRange = true
			}
		}()

		pipelineID := parsePipelineIDFromQueueItem(item)
		if pipelineID == 0 {
			rlog.PErrorf(pipelineID, "queueManager: invalid queue item key: %s, failed to parse to pipelineID, remove this item", pipelineID)
			return q.doStopAndRemove(item, false)
		}

		// get pipeline
		q.lock.RLock()
		p := q.pipelineCaches[pipelineID]
		q.lock.RUnlock()
		if p == nil {
			// pipeline not exist, remove this invalid item, continue handle next pipeline inside the queue
			rlog.PWarnf(pipelineID, "queueManager: failed to handle pipeline inside queue, pipeline not exist, pop from pending queue")
			return q.doStopAndRemove(item, false)
		}

		// queue validate
		// it will cause `concurrent map read and map write` panic if `pipeline_caches` is not locked.
		q.lock.RLock()
		validateResult := q.validatePipeline(p)
		q.lock.RUnlock()
		if !validateResult.Success {
			q.emitEvent(p, PendingQueueValidate, validateResult.Reason, events.EventLevelWarning)
			// stopRange if queue is strict mode
			return q.IsStrictMode()
		}

		// precheck before run
		customKVsOfAOP := map[interface{}]interface{}{}
		ctx := aop.NewContextForPipeline(*p, aoptypes.TuneTriggerPipelineInQueuePrecheckBeforePop, customKVsOfAOP)
		_ = aop.Handle(ctx)
		checkResultI, ok := ctx.TryGet(apistructs.PipelinePreCheckResultContextKey)
		if !ok {
			// no result, log and wait for another retry
			stopRange = false
			q.emitEvent(p, PendingQueueValidate,
				"queue precheck missing result, waiting for retry",
				events.EventLevelNormal)
			return
		}
		checkResult, ok := checkResultI.(apistructs.PipelineQueueValidateResult)
		if !ok {
			// invalid result, log and wait for another retry
			q.emitEvent(p, PendingQueueValidate,
				fmt.Sprintf("queue precheck result type is not expected, detail: %#v", checkResult),
				events.EventLevelNormal)
			stopRange = false
			return
		}
		// check result
		if checkResult.IsFailed() {
			// not retry if retryOption is nil
			if checkResult.RetryOption == nil {
				q.emitEvent(p, FailedQueue,
					fmt.Sprintf("validate failed(no retry option), stop and remove from queue, reason: %s", checkResult.Reason),
					events.EventLevelWarning)
				// mark pipeline as failed
				q.emitEvent(p, FailedQueue,
					"mark pipeline as failed",
					events.EventLevelNormal)
				q.ensureMarkPipelineFailed(p)
				return q.doStopAndRemove(item)
			}
			// need retry, sleep specific time
			q.emitEvent(p, PendingQueueValidate,
				fmt.Sprintf("validate failed(need retry), waiting for retry(%dsec), reason: %s", checkResult.RetryOption.IntervalSecond, checkResult.Reason),
				events.EventLevelNormal)
			// judge whether need reRange before sleep
			if q.needReRangePendingQueue() {
				return true
			}
			time.Sleep(time.Second * time.Duration(checkResult.RetryOption.IntervalSecond))
			// according to queue mode, check next pipeline or skip
			return q.IsStrictMode()
		}
		// do pop
		q.emitEvent(p, SuccessQueue,
			"validate success, try pop now",
			events.EventLevelNormal)
		stopRange = q.doPop(item)
		return
	})
}

func (q *defaultQueue) doPop(item priorityqueue.Item) (stopRange bool) {
	// pop now
	poppedKey := q.eq.PopPendingKey(item.Key())
	// queue cannot pop item anymore
	if poppedKey == "" {
		stopRange = true
		return
	}
	// send popped signal to channel
	pipelineID, _ := strconv.ParseUint(item.Key(), 10, 64)
	ch, ok := q.doneChanByPipelineID[pipelineID]
	if ok {
		ch <- struct{}{}
		close(ch)
		delete(q.doneChanByPipelineID, pipelineID)
	}
	// according to queue mode, check next pipeline or not
	return q.IsStrictMode()
}

func (q *defaultQueue) doStopAndRemove(item priorityqueue.Item, stopRangeOpt ...bool) (stopRange bool) {
	// pop out from pending queue
	q.eq.PendingQueue().Remove(item.Key())

	// according to queue mode, check next pipeline or not
	stopRange = q.IsStrictMode()
	if len(stopRangeOpt) > 0 {
		stopRange = stopRangeOpt[0]
	}
	return stopRange
}

func (q *defaultQueue) ensureMarkPipelineFailed(p *spec.Pipeline) {
	p.Status = apistructs.PipelineStatusFailed

	_ = loop.New(loop.WithDeclineLimit(time.Second*10), loop.WithDeclineRatio(2)).Do(func() (abort bool, err error) {
		// status
		if err := q.dbClient.UpdatePipelineBaseStatus(p.ID, p.Status); err != nil {
			err = fmt.Errorf("failed to mark pipeline as failed: %d, err: %v", p.ID, err)
			logrus.Error(err)
			return false, err
		}
		return true, nil
	})
}

// emitEvent
func (q *defaultQueue) emitEvent(p *spec.Pipeline, reason string, message string, eType string) string {
	now := time.Now()
	se := apistructs.PipelineEvent{
		Reason:         reason,
		Message:        message,
		Source:         apistructs.PipelineEventSource{Component: EventComponentQueueManager},
		FirstTimestamp: now,
		LastTimestamp:  now,
		Count:          1,
		Type:           eType,
	}
	events.EmitPipelineStreamEvent(p.ID, []*apistructs.PipelineEvent{&se})
	msg := fmt.Sprintf("queueManager: queueID: %s, queueName: %s, pipelineID: %d, Type: %s, Reason: %s, Message: %s",
		q.ID(), q.pq.Name, p.ID, eType, reason, message)
	rlog.PDebugf(p.ID, msg)
	return msg
}
