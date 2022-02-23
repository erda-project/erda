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
	"time"

	"github.com/coreos/etcd/clientv3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
	"github.com/erda-project/erda/pkg/strutil"
)

type config struct {
	LeaderCanBeWorker bool         `file:"leader_can_be_worker" env:"LEADER_CAN_BE_WORKER" default:"true"`
	Worker            workerConfig `file:"worker"`
}

type workerConfig struct {
	Candidate     candidateWorkerConfig `file:"candidate"`
	EtcdKeyPrefix string                `file:"etcd_key_prefix"`
}

type candidateWorkerConfig struct {
	ThresholdToBecomeOfficial time.Duration `file:"threshold_to_become_official" env:"THRESHOLD_TO_BECOME_OFFICIAL_WORKER" default:"15s"`
}

type provider struct {
	Log        logs.Logger
	Cfg        *config
	Election   election.Interface `autowired:"etcd-election@pipeline"`
	EtcdClient *clientv3.Client

	lock                sync.Mutex
	leaderHandler       func(ctx context.Context)
	currentWorkers      []worker.Worker
	workerAddHandler    WorkerAddHandler
	workerDeleteHandler WorkerDeleteHandler
}

type (
	WorkerAddHandler    func(ctx context.Context, ev Event)
	WorkerDeleteHandler func(ctx context.Context, ev Event)
)

func (p *provider) AddCandidateWorker(ctx context.Context, w worker.Worker) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	// register to currentWorkers
	p.currentWorkers = append(p.currentWorkers, w)
	// put into etcd
	if err := p.putWorkerIntoEtcd(ctx, w, worker.Candidate); err != nil {
		return err
	}
	// candidate do not participate in elect until become official
	p.Election.SetNonVoter(true)
	// a goroutine to let candidate worker become (official) worker, maybe use time.After
	var werr error
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(p.Cfg.Worker.Candidate.ThresholdToBecomeOfficial):
			p.Election.SetNonVoter(false)
			werr = p.putWorkerIntoEtcd(ctx, w, worker.Official)
			return
		}
	}()
	return werr
}

func (p *provider) OnLeader(h func(ctx context.Context)) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.leaderHandler = h
}

func (p *provider) ListWorkers(ctx context.Context, workerTypes ...worker.Type) ([]worker.Worker, error) {
	var workers []worker.Worker
	for _, typ := range workerTypes {
		typeWorkers, err := p.listWorkersByType(ctx, typ)
		if err != nil {
			return nil, err
		}
		workers = append(workers, typeWorkers...)
	}
	return workers, nil
}

func (p *provider) GetWorker(ctx context.Context, workerID worker.ID) (worker.Worker, error) {
	getResp, err := p.EtcdClient.Get(ctx, p.makeEtcdOfficialWorkerKey(workerID))
	if err != nil {
		return nil, err
	}
	if getResp == nil || getResp.Count == 0 || len(getResp.Kvs) == 0 {
		return nil, fmt.Errorf("not found")
	}
	kv := getResp.Kvs[0]
	if kv == nil || len(kv.Value) == 0 {
		return nil, fmt.Errorf("not found")
	}
	w, err := worker.NewFromBytes(kv.Value)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func (p *provider) OnWorkerAdd(h WorkerAddHandler) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.workerAddHandler = h
}

func (p *provider) OnWorkerDelete(h WorkerDeleteHandler) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.workerDeleteHandler = h
}

func (p *provider) listWorkersByType(ctx context.Context, typ worker.Type) ([]worker.Worker, error) {
	getResp, err := p.EtcdClient.Get(ctx, p.makeEtcdWorkerKeyPrefix(typ), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	var workers []worker.Worker
	for _, kv := range getResp.Kvs {
		w, err := worker.NewFromBytes(kv.Value)
		if err != nil {
			return nil, err
		}
		workers = append(workers, w)
	}
	return workers, nil
}

func (p *provider) makeEtcdWorkerKeyPrefix(typ worker.Type) string {
	return filepath.Join(p.Cfg.Worker.EtcdKeyPrefix, string(typ))
}

func (p *provider) getWorkerIDFromEtcdKey(key string, typ worker.Type) worker.ID {
	return worker.ID(strutil.TrimPrefixes(key, p.makeEtcdWorkerKeyPrefix(typ)))
}

// $prefix/worker/candidate/$workerID
func (p *provider) makeEtcdCandidateWorkerKey(workerID worker.ID) string {
	return filepath.Join(p.Cfg.Worker.EtcdKeyPrefix, worker.Candidate.String(), workerID.String())
}

// $prefix/worker/official/$workerID
func (p *provider) makeEtcdOfficialWorkerKey(workerID worker.ID) string {
	return filepath.Join(p.Cfg.Worker.EtcdKeyPrefix, worker.Official.String(), workerID.String())
}

// $prefix/worker/dispatch/$workerID/$logicTaskID(such as: pipelineID)
func (p *provider) makeEtcdWorkerTaskDispatchListenPrefix(workerID worker.ID) string {
	return filepath.Join(p.Cfg.Worker.EtcdKeyPrefix, "worker/dispatch", workerID.String())
}

// $prefix/worker/dispatch/$workerID/$logicTaskID(such as: pipelineID)
func (p *provider) makeEtcdWorkerTaskDispatchKey(workerID worker.ID, logicTaskID string) string {
	prefix := p.makeEtcdWorkerTaskDispatchListenPrefix(workerID)
	return filepath.Join(prefix, logicTaskID)
}

func (p *provider) putWorkerIntoEtcd(ctx context.Context, w worker.Worker, typ worker.Type) error {
	var key string
	switch typ {
	case worker.Official:
		key = p.makeEtcdOfficialWorkerKey(w.ID())
	case worker.Candidate:
		key = p.makeEtcdCandidateWorkerKey(w.ID())
	default:
		panic(fmt.Errorf("invalid worker type: %s", typ))
	}
	workerBytes, err := w.MarshalJSON()
	if err != nil {
		return err
	}
	if _, err := p.EtcdClient.Put(ctx, key, string(workerBytes)); err != nil {
		return err
	}
	return nil
}

func (p *provider) Init(ctx servicehub.Context) error {
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, options ...interface{}) interface{} {
	return p
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
