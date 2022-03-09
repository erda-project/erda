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
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"

	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
)

type Interface interface {
	RegisterCandidateWorker(ctx context.Context, w worker.Worker) error
	ListWorkers(ctx context.Context, workerTypes ...worker.Type) ([]worker.Worker, error)
	LeaderHandlerOnWorkerAdd(WorkerAddHandler)
	LeaderHandlerOnWorkerDelete(WorkerDeleteHandler)
	WorkerHandlerOnWorkerDelete(WorkerDeleteHandler)
	OnLeader(func(context.Context))
	AssignLogicTaskToWorker(ctx context.Context, workerID worker.ID, logicTask worker.LogicTask) error
	ListenPrefix(ctx context.Context, prefix string, putHandler, deleteHandler func(context.Context, *clientv3.Event))
	IsTaskBeingProcessed(ctx context.Context, logicTaskID worker.LogicTaskID) (bool, worker.ID)
	RegisterListener(l Listener)
	Start()
}

func (p *provider) RegisterCandidateWorker(ctx context.Context, w worker.Worker) error {
	p.Log.Infof("begin register candidate worker, workerID: %s", w.GetID())

	// check leader can be worker
	if p.Election.IsLeader() && !p.Cfg.Leader.IsWorker {
		p.Log.Warnf("leader cannot be worker, skip register candidate worker, workerID: %s", w.GetID())
		return nil
	}

	// check worker fields
	if err := p.checkWorkerFields(w); err != nil {
		p.Log.Errorf("failed to check worker fields, workerID: %s, err: %v", w.GetID(), err)
		return err
	}

	// register worker
	if err := p.registerWorker(ctx, w, worker.Candidate); err != nil {
		return err
	}

	p.lock.Lock()
	wctx, wcancel := context.WithCancel(ctx)
	p.forWorkerUse.myWorkers[w.GetID()] = workerWithCancel{Worker: w, Ctx: wctx, CancelFunc: wcancel}
	p.lock.Unlock()

	// promote to official
	go func() {
		p.promoteCandidateWorker(wctx, w)
		// begin listen after promoted
		p.workerListenIncomingLogicTask(wctx, w)
	}()

	// heartbeat report
	go p.workerContinueReportHeartbeat(wctx, w)

	// handle worker delete
	go p.workerHandleDelete(wctx, w)

	return nil
}

func (p *provider) ListWorkers(ctx context.Context, workerTypes ...worker.Type) ([]worker.Worker, error) {
	return p.listWorkers(ctx, workerTypes...)
}

func (p *provider) LeaderHandlerOnWorkerAdd(h WorkerAddHandler) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.started {
		panic(fmt.Errorf("cannot register LeaderHandlerOnWorkerAdd func after started"))
	}
	p.forLeaderUse.leaderHandlersOnWorkerAdd = append(p.forLeaderUse.leaderHandlersOnWorkerAdd, h)
}

func (p *provider) LeaderHandlerOnWorkerDelete(h WorkerDeleteHandler) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.started {
		panic(fmt.Errorf("cannot register LeaderHandlerOnWorkerDelete func after started"))
	}
	p.forLeaderUse.leaderHandlersOnWorkerDelete = append(p.forLeaderUse.leaderHandlersOnWorkerDelete, h)
}

func (p *provider) WorkerHandlerOnWorkerDelete(h WorkerDeleteHandler) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.started {
		panic(fmt.Errorf("cannot register WorkerHandlerOnWorkerDelete func after started"))
	}
	p.forWorkerUse.workerHandlersOnWorkerDelete = append(p.forWorkerUse.workerHandlersOnWorkerDelete, h)
}

func (p *provider) AssignLogicTaskToWorker(ctx context.Context, workerID worker.ID, task worker.LogicTask) error {
	_, err := p.EtcdClient.Put(ctx, p.makeEtcdWorkerTaskDispatchKey(workerID, task.GetLogicID()), string(task.GetData()))
	if err != nil {
		return err
	}
	p.addToTaskWorkerAssignMap(task.GetLogicID(), workerID)
	return err
}

func (p *provider) OnLeader(h func(ctx context.Context)) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.started {
		panic(fmt.Errorf("cannot register OnLeader func after started"))
	}
	p.forLeaderUse.leaderHandlers = append(p.forLeaderUse.leaderHandlers, h)
}

func (p *provider) IsTaskBeingProcessed(ctx context.Context, logicTaskID worker.LogicTaskID) (bool, worker.ID) {
	if !p.Election.IsLeader() {
		panic(fmt.Errorf("non-leader cannot invoke IsTaskBeingProcessed"))
	}
	for {
		p.lock.Lock()
		if p.forLeaderUse.initialized {
			p.lock.Unlock()
			break
		}
		p.lock.Unlock()
		time.Sleep(p.Cfg.Worker.RetryInterval)
	}
	p.lock.Lock()
	defer p.lock.Unlock()
	workerID, ok := p.forLeaderUse.findWorkerByTask[logicTaskID]
	if !ok {
		return false, ""
	}
	// check valid worker
	_, workerExist := p.forLeaderUse.allWorkers[workerID]
	if !workerExist {
		return false, ""
	}
	return true, workerID
}

func (p *provider) RegisterListener(l Listener) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.started {
		panic(fmt.Errorf("cannot register listener after started"))
	}
	p.listeners = append(p.listeners, l)
}

func (p *provider) Start() {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.started {
		return
	}
	p.started = true
}
