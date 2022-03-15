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

package types

import (
	"context"

	"github.com/erda-project/erda-proto-go/core/pipeline/queue/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/providers/queuemanager/pkg/queue/snapshot"
)

// QueueManager manage all queues and related pipelines.
type QueueManager interface {
	IdempotentAddQueue(pq *apistructs.PipelineQueue) Queue
	QueryQueueUsage(pq *apistructs.PipelineQueue) *pb.QueueUsage
	PutPipelineIntoQueue(pipelineID uint64) (popCh <-chan struct{}, needRetryIfErr bool, err error)
	PopOutPipelineFromQueue(pipelineID uint64)
	BatchUpdatePipelinePriorityInQueue(pq *apistructs.PipelineQueue, pipelineIDs []uint64) error
	Stop()
	SendQueueToEtcd(queueID uint64)
	ListenInputQueueFromEtcd(ctx context.Context)
	SendUpdatePriorityPipelineIDsToEtcd(queueID uint64, pipelineIDS []uint64)
	ListenUpdatePriorityPipelineIDsFromEtcd(ctx context.Context)
	SendPopOutPipelineIDToEtcd(pipelineID uint64)
	ListenPopOutPipelineIDFromEtcd(ctx context.Context)
	snapshot.Snapshot
}
