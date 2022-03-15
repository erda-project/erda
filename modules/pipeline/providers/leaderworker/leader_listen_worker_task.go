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

func (p *provider) leaderListenOfficialWorkerChange(ctx context.Context) {
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
					for _, h := range p.forLeaderUse.handlersOnWorkerAdd {
						h := h
						go h(ctx, ev)
					}
				case mvccpb.DELETE:
					for _, h := range p.forLeaderUse.handlersOnWorkerDelete {
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

func (p *provider) leaderListenLogicTaskChange(ctx context.Context) {
	p.Log.Infof("begin listen logic task change")
	defer p.Log.Infof("end listen logic task change")
	prefix := p.makeEtcdWorkerGeneralDispatchPrefix()
	p.ListenPrefix(ctx, prefix,
		func(ctx context.Context, event *clientv3.Event) {
			key := string(event.Kv.Key)
			workerID := p.getWorkerIDFromIncomingKey(key)
			logicTaskID := p.getWorkerTaskLogicIDFromIncomingKey(workerID, key)
			p.addToTaskWorkerAssignMap(logicTaskID, workerID)
		},
		func(ctx context.Context, event *clientv3.Event) {
			key := string(event.Kv.Key)
			workerID := p.getWorkerIDFromIncomingKey(key)
			logicTaskID := p.getWorkerTaskLogicIDFromIncomingKey(workerID, key)
			p.removeFromTaskWorkerAssignMap(logicTaskID, workerID)
		},
	)
}

func (p *provider) leaderInitTaskWorkerAssignMap(ctx context.Context) {
	p.lock.Lock()
	p.forLeaderUse.initialized = false
	p.forLeaderUse.findWorkerByTask = make(map[worker.LogicTaskID]worker.ID)
	p.forLeaderUse.findTaskByWorker = make(map[worker.ID]map[worker.LogicTaskID]struct{})
	p.lock.Unlock()

outLoop:
	for {
		workers, err := p.listWorkers(ctx)
		if err != nil {
			p.Log.Errorf("failed to list workers for leaderInitTaskWorkerAssignMap(auto retry), err: %v", err)
			time.Sleep(p.Cfg.Leader.RetryInterval)
			continue
		}
		for _, w := range workers {
			tasks, err := p.listWorkerTasks(ctx, w.GetID())
			if err != nil {
				p.Log.Errorf("failed to list worker tasks for leaderInitTaskWorkerAssignMap(auto retry), workerID: %s, err: %v", w.GetID(), err)
				time.Sleep(p.Cfg.Leader.RetryInterval)
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

func (p *provider) listWorkerTasks(ctx context.Context, workerID worker.ID) ([]worker.LogicTask, error) {
	prefix := p.makeEtcdWorkerLogicTaskListenPrefix(workerID)
	resp, err := p.EtcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var tasks []worker.LogicTask
	for _, kv := range resp.Kvs {
		logicTaskID := p.getWorkerTaskLogicIDFromIncomingKey(workerID, string(kv.Key))
		logicTaskData := kv.Value
		task := worker.NewLogicTask(logicTaskID, logicTaskData)
		tasks = append(tasks, task)
	}
	return tasks, nil
}
