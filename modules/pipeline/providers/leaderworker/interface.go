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

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"

	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
)

type Interface interface {
	AddCandidateWorker(ctx context.Context, w worker.Worker) error
	ListWorkers(ctx context.Context, workerTypes ...worker.Type) ([]worker.Worker, error)
	GetWorker(ctx context.Context, workerID worker.ID) (worker.Worker, error)
	OnWorkerAdd(WorkerAddHandler)
	OnWorkerDelete(WorkerDeleteHandler)
	OnLeader(func(context.Context))
	ListenPrefix(ctx context.Context, prefix string, putHandler, deleteHandler func(context.Context, *clientv3.Event))
	Dispatch(ctx context.Context, w worker.Worker, taskID string, taskData []byte) error
}

// Event .
type Event struct {
	Type     mvccpb.Event_EventType
	WorkerID worker.ID
}

func (p *provider) Run(ctx context.Context) error {
	// leader
	p.Election.OnLeader(p.leaderFramework)
	// worker
	p.workerFramework(ctx)
	return nil
}

func (p *provider) Dispatch(ctx context.Context, w worker.Worker, taskID string, taskData []byte) error {
	_, err := p.EtcdClient.Put(ctx, p.makeEtcdWorkerTaskDispatchKey(w.ID(), taskID), string(taskData))
	return err
}

func (p *provider) workerFramework(ctx context.Context) {
	p.lock.Lock()
	var currentWorkers []worker.Worker
	copy(currentWorkers, p.currentWorkers)
	p.lock.Unlock()
	for _, w := range p.currentWorkers {
		p.ListenPrefix(ctx, p.makeEtcdWorkerTaskDispatchListenPrefix(w.ID()),
			func(ctx context.Context, event *clientv3.Event) {
				// key added, do logic
				w.Handle(ctx, event.Kv.Value)
				// delete task key means done
				// TODO handle delete error to avoid key leak
				p.EtcdClient.Delete(ctx, string(event.Kv.Key))
			},
			nil,
		)
	}
}

// framework use etcd to store data
func (p *provider) leaderFramework(ctx context.Context) {
	// leader
	go func() {
		p.leaderHandler(ctx)
	}()

	// worker
	notify := make(chan Event)
	go func() {
		for {
			select {
			case ev := <-notify:
				switch ev.Type {
				case mvccpb.PUT:
					if p.workerAddHandler != nil {
						p.workerAddHandler(ctx, ev)
						continue
					}
				case mvccpb.DELETE:
					if p.workerDeleteHandler != nil {
						p.workerDeleteHandler(ctx, ev)
						continue
					}
				}
			}
		}
	}()
	p.ListenPrefix(ctx, p.makeEtcdWorkerKeyPrefix(worker.Official),
		func(ctx context.Context, ev *clientv3.Event) {
			notify <- Event{Type: mvccpb.PUT, WorkerID: p.getWorkerIDFromEtcdKey(string(ev.Kv.Key), worker.Official)}
		},
		func(ctx context.Context, ev *clientv3.Event) {
			notify <- Event{Type: mvccpb.DELETE, WorkerID: p.getWorkerIDFromEtcdKey(string(ev.Kv.Key), worker.Official)}
		},
	)
}

func (p *provider) ListenPrefix(ctx context.Context, prefix string, putHandler, deleteHandler func(context.Context, *clientv3.Event)) {
	for func() bool {
		wctx, wcancel := context.WithCancel(ctx)
		defer wcancel()
		wch := p.EtcdClient.Watch(wctx, prefix, clientv3.WithPrefix())
		for {
			select {
			case <-ctx.Done():
				return false
			case resp, ok := <-wch:
				if !ok {
					return true
				} else if resp.Err() != nil {
					p.Log.Errorf("failed to watch etcd prefix %s, error: %v", prefix, resp.Err())
					return true
				}
				for _, ev := range resp.Events {
					if ev.Kv == nil {
						continue
					}
					switch ev.Type {
					case mvccpb.PUT:
						if putHandler != nil {
							putHandler(wctx, ev)
						}
					case mvccpb.DELETE:
						if deleteHandler != nil {
							deleteHandler(wctx, ev)
						}
					}
				}
			}
		}
	}() {
	}
}
