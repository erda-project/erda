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

// IdempotentAddQueue add to to manager idempotent.
func (mgr *defaultManager) IdempotentAddQueue(pq *apistructs.PipelineQueue) types.Queue {
	mgr.qLock.Lock()
	defer mgr.qLock.Unlock()

	// construct newQueue first for later use
	newQueue := queue.New(pq,mgr.pluginsManage, queue.WithDBClient(mgr.dbClient))

	_, ok := mgr.queueByID[newQueue.ID()]
	if ok {
		// update queue
		mgr.queueByID[newQueue.ID()].Update(pq)
		return mgr.queueByID[newQueue.ID()]
	}

	// not exist, add new queue and start
	mgr.queueByID[newQueue.ID()] = newQueue
	qStopCh := make(chan struct{})
	mgr.queueStopChanByID[newQueue.ID()] = qStopCh
	newQueue.Start(qStopCh)

	return newQueue
}
