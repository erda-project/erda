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
	AssignLogicTaskToWorker(ctx context.Context, workerID worker.ID, logicTask worker.Tasker) error
	ListenPrefix(ctx context.Context, prefix string, putHandler, deleteHandler func(context.Context, *clientv3.Event))
	IsTaskBeingProcessed(ctx context.Context, logicTaskID worker.LogicTaskID) (bool, worker.ID)
}

func (p *provider) RegisterCandidateWorker(ctx context.Context, w worker.Worker) error {
	p.Log.Infof("begin register candidate worker, workerID: %s", w.GetID())

	// check leader can be worker
	if p.Election.IsLeader() && !p.Cfg.Leader.AlsoBeWorker {
		p.Log.Warnf("leader cannot be worker, skip register candidate worker, workerID: %s", w.GetID())
		return nil
	}

	// check worker fields
	if err := p.checkWorkerFields(w); err != nil {
		p.Log.Errorf("failed to check worker fields, workerID: %s, err: %v", w.GetID(), err)
		return err
	}

	// notify for worker-add
	if err := p.notifyWorkerAdd(ctx, w, worker.Candidate); err != nil {
		return err
	}

	p.lock.Lock()
	wctx, wcancel := context.WithCancel(ctx)
	p.workerUse.myWorkers[w.GetID()] = workerWithCancel{Worker: w, Ctx: wctx, CancelFunc: wcancel}
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
	p.leaderUse.leaderHandlersOnWorkerAdd = append(p.leaderUse.leaderHandlersOnWorkerAdd, h)
}

func (p *provider) LeaderHandlerOnWorkerDelete(h WorkerDeleteHandler) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.leaderUse.leaderHandlersOnWorkerDelete = append(p.leaderUse.leaderHandlersOnWorkerDelete, h)
}

func (p *provider) WorkerHandlerOnWorkerDelete(h WorkerDeleteHandler) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.workerUse.workerHandlersOnWorkerDelete = append(p.workerUse.workerHandlersOnWorkerDelete, h)
}

func (p *provider) AssignLogicTaskToWorker(ctx context.Context, workerID worker.ID, task worker.Tasker) error {
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
	p.leaderUse.leaderHandlers = append(p.leaderUse.leaderHandlers, h)
}

func (p *provider) IsTaskBeingProcessed(ctx context.Context, logicTaskID worker.LogicTaskID) (bool, worker.ID) {
	if !p.Election.IsLeader() {
		panic(fmt.Errorf("non-leader cannot invoke IsTaskBeingProcessed"))
	}
	for {
		p.lock.Lock()
		if p.leaderUse.initialized {
			p.lock.Unlock()
			break
		}
		p.lock.Unlock()
		time.Sleep(p.Cfg.Worker.RetryInterval)
	}
	p.lock.Lock()
	workerID, ok := p.leaderUse.findWorkerByTask[logicTaskID]
	p.lock.Unlock()
	return ok, workerID
}
