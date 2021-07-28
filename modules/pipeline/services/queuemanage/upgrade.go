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

package queuemanage

import (
	"fmt"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/pipengine/reconciler"
)

// UpdatePipelineQueue update queue all fields by id.
func (qm *QueueManage) BatchUpgradePipelinePriority(queue *apistructs.PipelineQueue, pipelineIDs []uint64, reconciler *reconciler.Reconciler) error {
	for _, i := range pipelineIDs {
		if !pipelineInpending(queue, i) {
			return fmt.Errorf("Pipeline %v is not in pending queue %v", i, queue.ID)
		}
	}

	var maxPriority int64
	for _, i := range queue.Usage.PendingDetails {
		if i.Priority > maxPriority {
			maxPriority = i.Priority
		}
	}

	if err := reconciler.QueueManager.UpdatePipelinePriorityInQueue(queue.ID, pipelineIDs, maxPriority); err != nil {
		return err
	}
	return nil
}

func pipelineInpending(queue *apistructs.PipelineQueue, pipelineID uint64) bool {
	for _, i := range queue.Usage.PendingDetails {
		if i.PipelineID == pipelineID {
			return true
		}
	}
	return false
}
