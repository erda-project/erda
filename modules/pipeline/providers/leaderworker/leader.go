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
	p.Log.Infof("leader framework starting ...")
	defer p.Log.Infof("leader framework stopped")

	for {
		p.lock.Lock()
		started := p.started
		p.lock.Unlock()
		if !started {
			p.Log.Warnf("waiting started")
			time.Sleep(p.Cfg.Leader.RetryInterval)
			continue
		}
		p.Log.Infof("leader framework started")
		break
	}

	// init before begin listen worker-task-change
	p.listeners = append([]Listener{
		&DefaultListener{BeforeExecOnLeaderFunc: p.initTaskWorkerAssignMap},
		&DefaultListener{BeforeExecOnLeaderFunc: asyncWrapper(p.listenOfficialWorkerChange)},
		&DefaultListener{BeforeExecOnLeaderFunc: asyncWrapper(p.listenWorkerTaskIncoming)},
		&DefaultListener{BeforeExecOnLeaderFunc: asyncWrapper(p.listenWorkerTaskDone)},
	}, p.listeners...)

	p.forLeaderUse.leaderHandlers = append(p.forLeaderUse.leaderHandlers, p.continueCleanup, p.workerLivenessProber)
	p.forWorkerUse.workerHandlersOnWorkerDelete = append(p.forWorkerUse.workerHandlersOnWorkerDelete, p.workerIntervalCleanupOnDelete)

	// before exec on leader
	for _, l := range p.listeners {
		l.BeforeExecOnLeader(ctx)
	}

	// exec on leader
	for _, h := range p.forLeaderUse.leaderHandlers {
		h := h
		go h(ctx)
	}

	// after exec on leader
	for _, l := range p.listeners {
		l.AfterExecOnLeader(ctx)
	}

	select {
	case <-ctx.Done():
		return
	}
}

func asyncWrapper(f func(ctx context.Context)) func(ctx context.Context) {
	return func(ctx context.Context) {
		go func() { f(ctx) }()
	}
}

func (p *provider) listenOfficialWorkerChange(ctx context.Context) {
	p.Log.Infof("begin listen official worker change")
	defer p.Log.Infof("end listen official worker change")
	// worker
	notify := make(chan Event)
	go func() {
		for {
			select {
			case ev := <-notify:
				switch ev.Type {
				case mvccpb.PUT:
					for _, h := range p.forLeaderUse.leaderHandlersOnWorkerAdd {
						h := h
						go h(ctx, ev)
					}
				case mvccpb.DELETE:
					for _, h := range p.forLeaderUse.leaderHandlersOnWorkerDelete {
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
			workerID := p.getWorkerIDFromEtcdWorkerKey(string(ev.Kv.Key), worker.Official)
			p.leaderUseDeleteInvalidWorker(workerID)
			logicTaskIDs, err := p.getWorkerLogicTaskIDs(ctx, workerID)
			if err == nil {
				notify <- Event{Type: mvccpb.DELETE, WorkerID: workerID, LogicTaskIDs: logicTaskIDs}
				p.leaderUseDeleteWorkerTaskAssign(workerID)
				return
			}
			// send notify of worker delete directly, otherwise new tasks will assign to this deleted worker at client-side(such like: dispatcher)
			notify <- Event{Type: mvccpb.DELETE, WorkerID: workerID, LogicTaskIDs: nil}

			// retry send notify of associated logic tasks
			p.Log.Errorf("failed to list worker tasks while worker deleted(auto retry), workerID: %s, err: %v", workerID, err)
			go func() {
				for {
					logicTaskIDs, err := p.getWorkerLogicTaskIDs(ctx, workerID)
					if err == nil {
						notify <- Event{Type: mvccpb.DELETE, WorkerID: workerID, LogicTaskIDs: logicTaskIDs}
						p.leaderUseDeleteWorkerTaskAssign(workerID)
						return
					}
					p.Log.Errorf("failed to retry list worker tasks(auto retry), workerID: %s, err: %v", workerID, err)
					time.Sleep(p.Cfg.Worker.RetryInterval)
				}
			}()
		},
	)
}

func (p *provider) getWorkerLogicTaskIDs(ctx context.Context, workerID worker.ID) ([]worker.LogicTaskID, error) {
	logicTasks, err := p.listWorkerTasks(ctx, workerID)
	if err == nil {
		logicTaskIDMap := make(map[worker.LogicTaskID]struct{})
		for _, task := range logicTasks {
			logicTaskIDMap[task.GetLogicID()] = struct{}{}
		}
		p.lock.Lock()
		for logicTaskID := range p.forLeaderUse.findTaskByWorker[workerID] {
			logicTaskIDMap[logicTaskID] = struct{}{}
		}
		p.lock.Unlock()
		var logicTaskIDs []worker.LogicTaskID
		for logicTaskID := range logicTaskIDMap {
			logicTaskIDs = append(logicTaskIDs, logicTaskID)
		}
		return logicTaskIDs, nil
	}
	return nil, err
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
		if _, ok := p.forLeaderUse.allWorkers[workerID]; !ok {
			danglingWorkerHeartbeatKeys = append(danglingWorkerHeartbeatKeys, string(kv.Key))
		}
	}
	p.lock.Unlock()
	for _, key := range danglingWorkerHeartbeatKeys {
		if _, err := p.EtcdClient.Delete(ctx, key); err != nil {
			p.Log.Errorf("failed to delete dangling worker heartbeat key(auto retry), key: %s, err: %v", key, err)
		}
	}

	// dangling dispatch key
	getResp, err = p.EtcdClient.Get(ctx, p.makeEtcdWorkerGeneralDispatchPrefix(), clientv3.WithPrefix())
	if err != nil {
		p.Log.Errorf("failed to prefix get worker task dispatch keys, err: %v", err)
		return
	}
	var danglingWorkerTaskDispatchKeys []string
	p.lock.Lock()
	for _, kv := range getResp.Kvs {
		workerID := p.getWorkerIDFromIncomingKey(string(kv.Key))
		if _, ok := p.forLeaderUse.allWorkers[workerID]; !ok {
			danglingWorkerTaskDispatchKeys = append(danglingWorkerTaskDispatchKeys, string(kv.Key))
		}
	}
	p.lock.Unlock()
	for _, key := range danglingWorkerTaskDispatchKeys {
		if _, err := p.EtcdClient.Delete(ctx, key); err != nil {
			p.Log.Errorf("failed to delete dangling worker task dispatch key(auto retry), key: %s, err: %v", key, err)
		}
	}
}

func (p *provider) listenWorkerTaskDone(ctx context.Context) {
	p.Log.Infof("begin listen worker task done")
	defer p.Log.Infof("end listen worker task done")
	prefix := p.makeEtcdWorkerGeneralDispatchPrefix()
	p.ListenPrefix(ctx, prefix,
		nil,
		func(ctx context.Context, event *clientv3.Event) {
			key := string(event.Kv.Key)
			workerID := p.getWorkerIDFromIncomingKey(key)
			logicTaskID := p.getWorkerTaskLogicIDFromIncomingKey(workerID, key)
			p.removeFromTaskWorkerAssignMap(logicTaskID, workerID)
		},
	)
}

func (p *provider) listenWorkerTaskIncoming(ctx context.Context) {
	p.Log.Infof("begin listen worker task incoming")
	defer p.Log.Infof("end listen worker task incoming")
	prefix := p.makeEtcdWorkerGeneralDispatchPrefix()
	p.ListenPrefix(ctx, prefix,
		func(ctx context.Context, event *clientv3.Event) {
			key := string(event.Kv.Key)
			workerID := p.getWorkerIDFromIncomingKey(key)
			logicTaskID := p.getWorkerTaskLogicIDFromIncomingKey(workerID, key)
			p.addToTaskWorkerAssignMap(logicTaskID, workerID)
		},
		nil,
	)
}

func (p *provider) initTaskWorkerAssignMap(ctx context.Context) {
	p.lock.Lock()
	p.forLeaderUse.initialized = false
	p.lock.Unlock()

	p.forLeaderUse.findWorkerByTask = make(map[worker.LogicTaskID]worker.ID)
	p.forLeaderUse.findTaskByWorker = make(map[worker.ID]map[worker.LogicTaskID]struct{})

outLoop:
	for {
		workers, err := p.listWorkers(ctx)
		if err != nil {
			p.Log.Errorf("failed to list workers for initTaskWorkerAssignMap(auto retry), err: %v", err)
			time.Sleep(p.Cfg.Worker.RetryInterval)
			continue
		}
		for _, w := range workers {
			tasks, err := p.listWorkerTasks(ctx, w.GetID())
			if err != nil {
				p.Log.Errorf("failed to list worker tasks for initTaskWorkerAssignMap(auto retry), workerID: %s, err: %v", w.GetID(), err)
				time.Sleep(p.Cfg.Worker.RetryInterval)
				continue outLoop
			}
			p.lock.Lock()
			p.forLeaderUse.findTaskByWorker[w.GetID()] = make(map[worker.LogicTaskID]struct{}, len(tasks))
			p.lock.Unlock()
			for _, task := range tasks {
				p.addToTaskWorkerAssignMap(task.GetLogicID(), w.GetID())
			}
		}
		// return if no error
		p.lock.Lock()
		p.forLeaderUse.initialized = true
		p.lock.Unlock()
		break
	}
}
