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
	"path/filepath"
	"reflect"
	"sync"

	"github.com/coreos/etcd/clientv3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
)

type provider struct {
	Log        logs.Logger
	Cfg        *config
	Election   election.Interface `autowired:"etcd-election@leader-worker"`
	EtcdClient *clientv3.Client

	lock sync.Mutex

	leaderUse leaderUse
	workerUse workerUse
}

type leaderUse struct {
	allWorkers map[worker.ID]worker.Worker

	initialized      bool
	findWorkerByTask map[worker.LogicTaskID]worker.ID
	findTaskByWorker map[worker.ID]map[worker.LogicTaskID]struct{}

	leaderHandlers               []func(ctx context.Context)
	leaderHandlersOnWorkerAdd    []WorkerAddHandler
	leaderHandlersOnWorkerDelete []WorkerDeleteHandler
}
type workerUse struct {
	myWorkers map[worker.ID]workerWithCancel

	workerHandlersOnWorkerDelete []WorkerDeleteHandler
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.leaderUse.allWorkers = make(map[worker.ID]worker.Worker)
	p.workerUse.myWorkers = make(map[worker.ID]workerWithCancel)
	if len(p.Cfg.Worker.EtcdKeyPrefixWithSlash) == 0 {
		return fmt.Errorf("failed to find config: worker.etcd_key_prefix_with_slash")
	}
	p.Cfg.Worker.EtcdKeyPrefixWithSlash = filepath.Clean(p.Cfg.Worker.EtcdKeyPrefixWithSlash) + "/"

	return nil
}

func (p *provider) addToTaskWorkerAssignMap(logicTaskID worker.LogicTaskID, workerID worker.ID) {
	p.lock.Lock()
	defer p.lock.Unlock()
	// findWorkerByTask
	p.leaderUse.findWorkerByTask[logicTaskID] = workerID
	// findTaskByWorker
	if p.leaderUse.findTaskByWorker[workerID] == nil {
		p.leaderUse.findTaskByWorker[workerID] = make(map[worker.LogicTaskID]struct{})
	}
	p.leaderUse.findTaskByWorker[workerID][logicTaskID] = struct{}{}
}

func (p *provider) removeFromTaskWorkerAssignMap(logicTaskID worker.LogicTaskID, workerID worker.ID) {
	p.lock.Lock()
	defer p.lock.Unlock()
	// findWorkerByTask
	delete(p.leaderUse.findWorkerByTask, logicTaskID)
	// findTaskByWorker
	workerTasks, ok := p.leaderUse.findTaskByWorker[workerID]
	if ok {
		delete(workerTasks, logicTaskID)
	}
}

func (p *provider) Run(ctx context.Context) error {
	p.Election.OnLeader(p.leaderFramework)
	p.Election.OnLeader(p.workerLivenessProber)
	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("leader-worker", &servicehub.Spec{
		Services:     []string{"leader-worker"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: []string{"etcd-election"},
		Description:  "pipeline-level leader&worker",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
