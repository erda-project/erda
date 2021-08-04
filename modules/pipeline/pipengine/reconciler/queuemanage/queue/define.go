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
	"sync"

	"github.com/erda-project/erda/modules/pipeline/spec"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/enhancedqueue"
)

// defaultQueue is used to implement Queue.
type defaultQueue struct {
	// pq is original pipeline queue.
	pq *apistructs.PipelineQueue

	// eq is enhanced priority queue, transfer from pq.
	eq *enhancedqueue.EnhancedQueue

	// doneChannels
	doneChanByPipelineID map[uint64]chan struct{}

	// pipeline caches
	pipelineCaches map[uint64]*spec.Pipeline

	// dbClient
	dbClient *dbclient.Client

	lock sync.RWMutex

	// started represents queue started handle process
	started bool

	// ranging about
	rangingPendingQueue         bool
	needReRangePendingQueueFlag bool
	currentItemKeyAtRanging     string // is meaningful only when rangingPendingQueue is true
	rangeAtOnceCh               chan bool

	// is updating pending queue
	updatingPendingQueue bool
}

func New(pq *apistructs.PipelineQueue, ops ...Option) *defaultQueue {
	newQueue := defaultQueue{
		pq:                   pq,
		eq:                   enhancedqueue.NewEnhancedQueue(pq.Concurrency),
		doneChanByPipelineID: make(map[uint64]chan struct{}),
		pipelineCaches:       make(map[uint64]*spec.Pipeline),
		rangeAtOnceCh:        make(chan bool),
	}

	// apply options
	for _, op := range ops {
		op(&newQueue)
	}

	return &newQueue
}

type Option func(*defaultQueue)

func WithDBClient(dbClient *dbclient.Client) Option {
	return func(q *defaultQueue) {
		q.dbClient = dbClient
	}
}

func (q *defaultQueue) ID() string {
	return strconv.FormatUint(q.pq.ID, 10)
}

func (q *defaultQueue) needReRangePendingQueue() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.needReRangePendingQueueFlag
}

func (q *defaultQueue) unsetNeedReRangePendingQueueFlag() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.needReRangePendingQueueFlag = false
}

func (q *defaultQueue) setIsRangingPendingQueueFlag() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.rangingPendingQueue = true
}

func (q *defaultQueue) unsetIsRangingPendingQueueFlag() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.rangingPendingQueue = false
}

func (q *defaultQueue) setCurrentItemKeyAtRanging(itemKey string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.currentItemKeyAtRanging = itemKey
}

func (q *defaultQueue) getIsRangingPendingQueue() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.rangingPendingQueue
}

func (q *defaultQueue) setIsUpdatingPendingQueueFlag() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.updatingPendingQueue = true
}

func (q *defaultQueue) unsetIsUpdatingPendingQueueFlag() {
	q.updatingPendingQueue = false
}

func (q *defaultQueue) getIsUpdatingPendingQueue() bool {
	q.lock.RLock()
	defer q.lock.RUnlock()
	return q.rangingPendingQueue
}
