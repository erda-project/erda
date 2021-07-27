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

package types

import (
	queuepb "github.com/erda-project/erda-proto-go/core/pipeline/queue/pb"
)

// QueueManager manage all queues and related pipelines.
type QueueManager interface {
	IdempotentAddQueue(pq *queuepb.Queue) Queue
	QueryQueueUsage(pq *queuepb.Queue) *queuepb.QueueUsage
	PutPipelineIntoQueue(pipelineID uint64) (popCh <-chan struct{}, needRetryIfErr bool, err error)
	PopOutPipelineFromQueue(pipelineID uint64)
}
