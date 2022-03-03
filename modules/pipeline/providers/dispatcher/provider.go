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
	"reflect"
	"sync"

	"github.com/buraksezer/consistent"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	"github.com/erda-project/erda/modules/pipeline/dbclient"
	"github.com/erda-project/erda/modules/pipeline/providers/leaderworker"
	"github.com/erda-project/erda/modules/pipeline/providers/reconciler"
)

type provider struct {
	Log        logs.Logger
	Cfg        *config
	Lw         leaderworker.Interface
	Reconciler reconciler.Interface

	Mysql    mysqlxorm.Interface
	dbClient *dbclient.Client

	pipelineIDsChan chan uint64
	consistent      *consistent.Consistent

	lock           sync.Mutex
	dispatchingIDs sync.Map
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.pipelineIDsChan = make(chan uint64, p.Cfg.Concurrency)
	p.dbClient = &dbclient.Client{Engine: p.Mysql.DB()}
	c, err := p.makeConsistent(ctx)
	if err != nil {
		return err
	}
	p.consistent = c

	return nil
}

func (p *provider) Run(ctx context.Context) error {
	// just register handler, and leader-worker provider will handle properly
	p.Lw.OnLeader(p.continueDispatcher)
	p.Lw.LeaderHandlerOnWorkerAdd(p.onWorkerAdd)
	p.Lw.LeaderHandlerOnWorkerDelete(p.onWorkerDelete)
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
