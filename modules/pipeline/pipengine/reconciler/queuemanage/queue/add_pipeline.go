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
	"strconv"
	"time"

	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/priorityqueue"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/rlog"
	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (q *defaultQueue) AddPipelineIntoQueue(p *spec.Pipeline, doneCh chan struct{}) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// make item key by pipeline info
	itemKey := makeItemKey(p)
	// set priority
	priority := q.pq.Priority
	// use pipeline level priority if have
	if p.Extra.QueueInfo != nil && p.Extra.QueueInfo.CustomPriority > 0 {
		priority = p.Extra.QueueInfo.CustomPriority
	}
	// createdTime
	createdTime := p.TimeCreated
	if createdTime == nil {
		now := time.Now()
		createdTime = &now
	}

	q.eq.Add(itemKey, priority, *createdTime)
	q.doneChanByPipelineID[p.ID] = doneCh
}

// parsePipelineIDFromQueueItem
// item key is the pipeline id
func parsePipelineIDFromQueueItem(item priorityqueue.Item) uint64 {
	pipelineID, err := strconv.ParseUint(item.Key(), 10, 64)
	if err != nil {
		rlog.Errorf("failed to parse pipeline id from queue item key, key: %s, err: %v", item.Key(), err)
		return 0
	}
	return pipelineID
}

// makeItemKey
func makeItemKey(p *spec.Pipeline) string {
	return strconv.FormatUint(p.ID, 10)
}
