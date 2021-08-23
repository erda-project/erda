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
	newQueue := queue.New(pq, queue.WithDBClient(mgr.dbClient))

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
