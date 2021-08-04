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
	"fmt"

	"github.com/erda-project/erda/modules/pipeline/spec"
)

func (q *defaultQueue) BatchUpdatePipelinePriorityInQueue(pipelines []*spec.Pipeline) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.updatingPendingQueue = true
	defer q.unsetIsUpdatingPendingQueueFlag()

	peeked := q.eq.PendingQueue().Peek()
	if peeked == nil {
		return fmt.Errorf("pending queue is empty")
	}

	priority := peeked.Priority() + int64(len(pipelines))

	for _, p := range pipelines {
		pipelineID := makeItemKey(p)
		if !q.eq.InPending(pipelineID) {
			return fmt.Errorf("failed to query pipeline: %s in pending queue", pipelineID)
		}
	}

	for _, p := range pipelines {
		if p.Extra.QueueInfo.PriorityChangeHistory == nil {
			p.Extra.QueueInfo.PriorityChangeHistory = []int64{p.Extra.QueueInfo.CustomPriority}
		}
		p.Extra.QueueInfo.PriorityChangeHistory = append(p.Extra.QueueInfo.PriorityChangeHistory, priority)
		p.Extra.QueueInfo.CustomPriority = priority
		priority -= 1
		if err := q.dbClient.UpdatePipelineExtraByPipelineID(p.ID, &p.PipelineExtra); err != nil {
			return err
		}
		q.addPipelineIntoQueueUnblock(p, nil)
	}

	return nil
}
