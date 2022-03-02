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
	"strconv"
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
	Worker workerConfig `file:"worker"`
}
type workerConfig struct {
	RetryInterval time.Duration `file:"retry_interval" default:"5s"`
}

type provider struct {
	Log logs.Logger
	Cfg *config

	// inject
	MySQLXOrm    mysqlxorm.Interface
	QueueManager queuemanager.Interface
	Dispatcher   dispatcher.Interface
	Reconciler   reconciler.Interface
	Lw           leaderworker.Interface

	// manual
	dbClient          *dbclient.Client
	actionExecutorMgr *actionexecutor.Manager
}

func (p *provider) Init(ctx servicehub.Context) error {
	// dbclient
	p.dbClient = &dbclient.Client{Engine: p.MySQLXOrm.DB()}

	// set bundle before initialize scheduler, because scheduler need use bdl get clusters
	bdl := bundle.New(bundle.WithAllAvailableClients(), bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second, time.Second))))
	clusterinfo.Initialize(bdl)

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

func (p *provider) Run(ctx context.Context) error {
	err := p.Lw.RegisterCandidateWorker(ctx, worker.New(worker.WithHandler(p.reconcilePipeline)))
	if err != nil {
		return err
	}
	p.Lw.WorkerHandlerOnWorkerDelete(func(ctx context.Context, ev leaderworker.Event) {
		for {
			err := p.Lw.RegisterCandidateWorker(ctx, worker.New(worker.WithHandler(p.reconcilePipeline)))
			if err == nil {
				return
			}
			p.Log.Errorf("failed to add new candidate worker when old worker deleted(auto retry), old workerID: %s, err: %v", ev.WorkerID, err)
			time.Sleep(p.Cfg.Worker.RetryInterval)
		}
	})

	return nil
}

func (p *provider) reconcilePipeline(ctx context.Context, logicTask worker.Tasker) {
	if logicTask == nil {
		p.Log.Warnf("logicTask is nil, skip reconcile pipeline")
		return
	}
	idstr := logicTask.GetLogicID().String()
	pipelineID, err := strconv.ParseUint(idstr, 10, 64)
	if err != nil {
		p.Log.Errorf("failed to parse pipelineID from logicTask(no retry), logicTaskID: %s, err: %v", idstr, err)
		return
	}
	p.Reconciler.Reconcile(ctx, pipelineID)
}

func init() {
	interfaceType := reflect.TypeOf((*Interface)(nil)).Elem()
	servicehub.Register("pipengine", &servicehub.Spec{
		Services:     []string{"pipengine"},
		Types:        []reflect.Type{interfaceType},
		Dependencies: []string{""},
		Description:  "pipeline engine",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
