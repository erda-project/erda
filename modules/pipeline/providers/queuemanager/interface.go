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

package queuemanager

import (
	"context"
	"sync"
	"time"

	"github.com/erda-project/erda-proto-go/pipeline/pb"
	"github.com/erda-project/erda/apistructs"
)

type Interface interface {
	DistributedHandleIncomingPipeline(ctx context.Context, pipelineID uint64)
	DistributedStopPipeline(ctx context.Context, pipelineID uint64)
	DistributedQueryQueueUsage(ctx context.Context, queue *apistructs.PipelineQueue) *pb.QueueUsage
	DistributedUpdateQueue(ctx context.Context, queueID uint64)
	DistributedBatchUpdatePipelinePriority(ctx context.Context, queueID uint64, pipelineIDsOrderByPriorityFromHighToLow []uint64)
}

func (q *provider) DistributedHandleIncomingPipeline(ctx context.Context, pipelineID uint64) {
	var wg sync.WaitGroup
	// put into etcd
	go func() {
		wg.Add(1)
		defer wg.Done()
		key := makeIncomingPipelineListenKey(q.Cfg.IncomingPipelineCfg.ListenPrefixWithSlash, pipelineID)
		for {
			if _, err := q.EtcdClient.Put(ctx, key, ""); err != nil {
				q.Log.Warnf("failed to distribute incoming pipeline(auto retry), pipelineID: %d, err: %v", pipelineID, err)
				time.Sleep(q.Cfg.IncomingPipelineCfg.RetryInterval)
				continue
			}
			return
		}
	}()
	// put into mysql
	go func() {
		wg.Add(1)
		defer wg.Done()
		for {
			if err := q.dbClient.UpdatePipelineBaseStatus(pipelineID, apistructs.PipelineStatusQueue); err != nil {
				q.Log.Warnf("failed to distribute incoming pipeline(auto retry), pipelineID: %d, err: %v", pipelineID, err)
				time.Sleep(q.Cfg.IncomingPipelineCfg.RetryInterval)
				continue
			}
			return
		}
	}()

	wg.Wait()
	q.Log.Debugf("distribute incoming pipeline success, pipelineID: %d", pipelineID)
}

func (q *provider) DistributedStopPipeline(ctx context.Context, pipelineID uint64) {
	q.QueueManager.SendPopOutPipelineIDToEtcd(pipelineID)
}

func (q *provider) DistributedQueryQueueUsage(ctx context.Context, pq *apistructs.PipelineQueue) *pb.QueueUsage {
	return q.QueueManager.QueryQueueUsage(pq)
}

func (q *provider) DistributedUpdateQueue(ctx context.Context, queueID uint64) {
	q.QueueManager.SendQueueToEtcd(queueID)
}

func (q *provider) DistributedBatchUpdatePipelinePriority(ctx context.Context, queueID uint64, pipelineIDsOrderByPriorityFromHighToLow []uint64) {
	q.QueueManager.SendUpdatePriorityPipelineIDsToEtcd(queueID, pipelineIDsOrderByPriorityFromHighToLow)
}
