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

package engine

import (
	"context"
	"reflect"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/pipengine/actionexecutor"
	"github.com/erda-project/erda/modules/pipeline/pkg/clusterinfo"
	"github.com/erda-project/erda/modules/pipeline/providers/dispatcher"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker/worker"
	"github.com/erda-project/erda/modules/pipeline/providers/queuemanager"
	"github.com/erda-project/erda/modules/pipeline/providers/reconciler"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type config struct {
	RetryInterval time.Duration `file:"retry_interval" default:"5s"`

	Worker workerConfig `file:"worker"`
}
type workerConfig struct {
	RetryInterval time.Duration `file:"retry_interval" default:"5s"`
}

type provider struct {
	Log logs.Logger
	Cfg *config

	// inject
	MySQL        mysqlxorm.Interface
	QueueManager queuemanager.Interface
	Dispatcher   dispatcher.Interface
	Reconciler   reconciler.Interface
	LW           leaderworker.Interface

	// manual
	dbClient          *dbclient.Client
	actionExecutorMgr *actionexecutor.Manager
}

func (p *provider) Init(ctx servicehub.Context) error {
	// dbclient
	p.dbClient = &dbclient.Client{Engine: p.MySQL.DB()}

	// cluster info
	// TODO setup inside clusterinfo provider later
	p.initClusterInfoUntilSuccess(ctx)

	// action executor manager
	_, cfgChan, err := p.dbClient.ListPipelineConfigsOfActionExecutor()
	if err != nil {
		return err
	}
	mgr := actionexecutor.GetManager()
	p.actionExecutorMgr = mgr
	if err := mgr.Initialize(cfgChan); err != nil {
		return err
	}

	return nil
}

func (p *provider) initClusterInfoUntilSuccess(ctx context.Context) {
	bdl := bundle.New(bundle.WithAllAvailableClients(), bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second, time.Second))))
	clusterinfo.Initialize(bdl)

	// continuous register cluster hook
	// pipeline depends on eventbox, so if eventbox is under rebooting, pipeline will failed to register cluster hook
	for {
		err := clusterinfo.RegisterClusterHook()
		if err == nil {
			break
		}
		p.Log.Infof("failed to register cluster hook(auto retry), err: %v", err)
		time.Sleep(p.Cfg.RetryInterval)
		continue
	}
}

func (p *provider) Run(ctx context.Context) error {
	err := p.LW.RegisterCandidateWorker(ctx, worker.New(worker.WithHandler(p.reconcileOnePipeline)))
	if err != nil {
		return err
	}
	p.LW.WorkerHandlerOnWorkerDelete(p.workerHandlerOnWorkerDelete)
	p.LW.Start()

	return nil
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("pipengine", &servicehub.Spec{
		Services:     []string{"pipengine"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: nil,
		Description:  "pipeline engine",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
