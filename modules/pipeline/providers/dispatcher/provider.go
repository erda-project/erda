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

package dispatcher

import (
	"context"
	"path/filepath"
	"reflect"
	"sync"
	"time"

	"github.com/buraksezer/consistent"
	"github.com/coreos/etcd/clientv3"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/strutil"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
)

type config struct {
	EtcdKeyPrefix                  string        `file:"etcd_key_prefix" env:"DISPATCHER_ETCD_KEY_PREFIX" default:"/devops/pipeline/v2/dispatcher/"`
	Concurrency                    int           `file:"concurrency" default:"100"`
	IntervalOfLoadRunningPipelines time.Duration `file:"interval_of_load_running_pipelines" env:"INTERVAL_OF_LOAD_RUNNING_PIPELINES" default:"30s"`
}

type provider struct {
	Log        logs.Logger
	Cfg        *config
	EtcdClient *clientv3.Client
	Lw         leaderworker.Interface

	Mysql    mysqlxorm.Interface
	dbClient *dbclient.Client

	pipelineIDsChan chan uint64
	consistent      *consistent.Consistent

	lock sync.Mutex
}

func (p *provider) Init(ctx servicehub.Context) error {
	// cfg
	p.Cfg.EtcdKeyPrefix = filepath.Clean(p.Cfg.EtcdKeyPrefix) + "/"

	// consistent
	consistentCfg := consistent.Config{
		Hasher:            defaultHash{},
		PartitionCount:    7,
		ReplicationFactor: 2,
		Load:              1.25,
	}
	var consistentMembers []consistent.Member
	workers, err := p.Lw.ListWorkers(ctx, worker.Official)
	if err != nil {
		return err
	}
	for _, w := range workers {
		consistentMembers = append(consistentMembers, w)
	}
	p.consistent = consistent.New(consistentMembers, consistentCfg)
	p.pipelineIDsChan = make(chan uint64, p.Cfg.Concurrency)
	p.dbClient = &dbclient.Client{Engine: p.Mysql.DB()}

	return nil
}

func (p *provider) Dispatch(ctx context.Context, pipelineID uint64) {
	p.pipelineIDsChan <- pipelineID
}

func (p *provider) makeEtcdDispatchKey(pipelineID uint64) string {
	return p.Cfg.EtcdKeyPrefix + strutil.String(pipelineID)
}

func (p *provider) Provide(ctx servicehub.DependencyContext, options ...interface{}) interface{} {
	return p
}

func (p *provider) Run(ctx context.Context) error {
	// just register handler, and leaderworker provider will handle properly
	p.Lw.OnLeader(p.loadRunningPipelines)
	p.Lw.OnLeader(p.continueDispatcher)
	p.Lw.OnWorkerAdd(p.onWorkerAdd)
	p.Lw.OnWorkerDelete(p.onWorkerDelete)
	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("dispatcher", &servicehub.Spec{
		Services:     []string{"dispatcher"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: []string{"leader-worker"},
		Description:  "pipeline engine dispatcher",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
