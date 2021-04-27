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
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/erda-project/erda-proto-go/pipeline/pb"
	"github.com/erda-project/erda/modules/pipeline/pipengine/queue/priorityqueue"
	"github.com/erda-project/erda/pkg/numeral"
)

func (q *defaultQueue) Usage() pb.QueueUsage {
	q.lock.RLock()
	defer q.lock.RUnlock()

	// processing
	var (
		inUseCPU          float64
		inUseMemoryMB     float64
		processingDetails = make([]*pb.QueueUsageItem, 0)
	)
	q.eq.ProcessingQueue().Range(func(item priorityqueue.Item) (stopRange bool) {
		pipelineID := parsePipelineIDFromQueueItem(item)
		existP := q.pipelineCaches[pipelineID]
		if existP == nil {
			return false
		}
		resources := existP.GetPipelineAppliedResources()
		inUseCPU += resources.Requests.CPU
		inUseMemoryMB += resources.Requests.MemoryMB
		processingDetails = append(processingDetails, &pb.QueueUsageItem{
			PipelineID:       pipelineID,
			RequestsCPU:      resources.Requests.CPU,
			RequestsMemoryMB: resources.Requests.MemoryMB,
			Index:            int64(item.Index()),
			Priority:         item.Priority(),
			AddedTime:        timestamppb.New(item.CreationTime()),
		})
		return false
	})

	// pending
	var pendingDetails = make([]*pb.QueueUsageItem, 0)
	q.eq.PendingQueue().Range(func(item priorityqueue.Item) (stopRange bool) {
		pipelineID := parsePipelineIDFromQueueItem(item)
		existP := q.pipelineCaches[pipelineID]
		if existP == nil {
			return false
		}
		resources := existP.GetPipelineAppliedResources()
		pendingDetails = append(pendingDetails, &pb.QueueUsageItem{
			PipelineID:       pipelineID,
			RequestsCPU:      resources.Requests.CPU,
			RequestsMemoryMB: resources.Requests.MemoryMB,
			Index:            int64(item.Index()),
			Priority:         item.Priority(),
			AddedTime:        timestamppb.New(time.Now()),
		})
		return false
	})

	return pb.QueueUsage{
		InUseCPU:          inUseCPU,
		InUseMemoryMB:     inUseMemoryMB,
		RemainingCPU:      numeral.SubFloat64(q.pq.MaxCPU, inUseCPU),
		RemainingMemoryMB: numeral.SubFloat64(q.pq.MaxMemoryMB, inUseMemoryMB),
		ProcessingCount:   int64(len(processingDetails)),
		PendingCount:      int64(len(pendingDetails)),
		ProcessingDetails: processingDetails,
		PendingDetails:    pendingDetails,
	}
}
