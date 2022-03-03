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

package leaderworker

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
)

func (p *provider) leaderFramework(ctx context.Context) {
	p.Log.Infof("start leader framework")
	defer p.Log.Infof("end leader framework")

	p.leaderUse.leaderHandlers = append(p.leaderUse.leaderHandlers, p.continueCleanup)
	p.workerUse.workerHandlersOnWorkerDelete = append(p.workerUse.workerHandlersOnWorkerDelete, p.workerIntervalCleanupOnDelete)

	// leader
	for _, h := range p.leaderUse.leaderHandlers {
		h := h
		go h(ctx)
	}

	// worker
	notify := make(chan Event)
	go func() {
		for {
			select {
			case ev := <-notify:
				switch ev.Type {
				case mvccpb.PUT:
					for _, h := range p.leaderUse.leaderHandlersOnWorkerAdd {
						h := h
						go h(ctx, ev)
					}
				case mvccpb.DELETE:
					for _, h := range p.leaderUse.leaderHandlersOnWorkerDelete {
						h := h
						go h(ctx, ev)
					}
				}
			}
		}
	}()
	p.ListenPrefix(ctx, p.makeEtcdWorkerKeyPrefix(worker.Official),
		func(ctx context.Context, ev *clientv3.Event) {
			notify <- Event{Type: mvccpb.PUT, WorkerID: p.getWorkerIDFromEtcdWorkerKey(string(ev.Kv.Key), worker.Official)}
		},
		func(ctx context.Context, ev *clientv3.Event) {
			notify <- Event{Type: mvccpb.DELETE, WorkerID: p.getWorkerIDFromEtcdWorkerKey(string(ev.Kv.Key), worker.Official)}
		},
	)
}

func (p *provider) continueCleanup(ctx context.Context) {
	p.Log.Infof("begin continue cleanup")
	defer p.Log.Infof("end continue cleanup")
	ticket := time.NewTicker(p.Cfg.Leader.CleanupInterval)
	defer ticket.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticket.C:
			p.intervalCleanupDanglingKeysWithoutRetry(ctx)
		}
	}
}

func (p *provider) intervalCleanupDanglingKeysWithoutRetry(ctx context.Context) {
	// dangling heartbeat keys
	getResp, err := p.EtcdClient.Get(ctx, p.makeEtcdWorkerHeartbeatKeyPrefix(), clientv3.WithPrefix())
	if err != nil {
		p.Log.Errorf("failed to prefix get heartbeat keys, err: %v", err)
		return
	}
	var danglingWorkerHeartbeatKeys []string
	p.lock.Lock()
	for _, kv := range getResp.Kvs {
		workerID := p.getWorkerIDFromEtcdWorkerHeartbeatKey(string(kv.Key))
		if _, ok := p.leaderUse.allWorkers[workerID]; !ok {
			danglingWorkerHeartbeatKeys = append(danglingWorkerHeartbeatKeys, string(kv.Key))
		}
	}
	p.lock.Unlock()
	for _, key := range danglingWorkerHeartbeatKeys {
		_, _ = p.EtcdClient.Delete(ctx, key)
	}
}
