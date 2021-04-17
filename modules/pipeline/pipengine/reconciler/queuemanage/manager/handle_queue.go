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

package manager

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/queuemanage/queue"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler/queuemanage/types"
)

// HandleAllQueues handle queues in order.
func (mgr *defaultManager) HandleAllQueues() {
	mgr.qLock.Lock()
	defer mgr.qLock.Unlock()

	doneCh := make(chan struct{}, len(mgr.queueByID))

	// iterate all queues
	for _, q := range mgr.queueByID {

		// every queue handle in one goroutine
		go func(q types.Queue) {
			mgr.handleOneQueue(q)
			doneCh <- struct{}{}
		}(q)
	}

	for i := 0; i < len(mgr.queueByID); i++ {
		<-doneCh
	}
}

// handleOneQueue handle all pipelines inside this queue.
func (mgr *defaultManager) handleOneQueue(queue types.Queue) {
	queue.RangePendingQueue(mgr)
}

func (mgr *defaultManager) UpdatePipelineQueue(pq *apistructs.PipelineQueue) {
	mgr.qLock.Lock()
	defer mgr.qLock.Unlock()

	newQueue := queue.New(pq)

	_, ok := mgr.queueByID[newQueue.ID()]
	if !ok {
		mgr.queueByID[newQueue.ID()] = newQueue
	} else {
		mgr.queueByID[newQueue.ID()].Update(pq)
	}
}
