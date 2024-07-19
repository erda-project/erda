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

package testplan

import (
	"context"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	election "github.com/erda-project/erda-infra/providers/etcd-election"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-proto-go/core/dop/autotest/testplan/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/apps/dop/dao"
	"github.com/erda-project/erda/internal/apps/dop/providers/autotest/testplan/db"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/tools/pipeline/providers/reconciler/rutil"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

type config struct {
	ExecHistoryDBGCDuration time.Duration `file:"exec_history_dbgc_duration" env:"EXEC_HISTORY_DBGC_DURATION" default:"12h"`
	ExecHistoryRetainHour   time.Duration `file:"exec_history_retain_hour" env:"EXEC_HISTORY_RETAIN_HOUR" default:"720h"`
}

// +provider
type provider struct {
	Cfg      *config
	Log      logs.Logger
	Register transport.Register `autowired:"service-register" required:"true"`
	DB       *gorm.DB           `autowired:"mysql-client"`
	bundle   *bundle.Bundle

	TestPlanService *TestPlanService
	Org             org.Interface
	RouterManager   httpserver.RouterManager
	Election        election.Interface `autowired:"etcd-election@autotest-testplan-provider"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.bundle = bundle.New(bundle.WithErdaServer())

	p.TestPlanService = &TestPlanService{
		p: p,
		db: db.TestPlanDB{
			DBClient: &dao.DBClient{
				DBEngine: &dbengine.DBEngine{
					p.DB,
				},
			},
		},
		bdl: p.bundle,
		org: p.Org,
	}

	if p.Register != nil {
		pb.RegisterTestPlanServiceImp(p.Register, p.TestPlanService)
	}
	p.Election.OnLeader(p.ExecHistoryGC)
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	<-p.RouterManager.Started()
	if err := p.registerWebHook(); err != nil {
		return err
	}
	return nil
}

func (p *provider) ExecHistoryGC(ctx context.Context) {
	p.Log.Infof("start exec history gc")
	rutil.ContinueWorking(ctx, p.Log, func(ctx context.Context) rutil.WaitDuration {
		p.TestPlanService.doExecHistoryGC()
		return rutil.ContinueWorkingWithDefaultInterval
	}, rutil.WithContinueWorkingDefaultRetryInterval(p.Cfg.ExecHistoryDBGCDuration))
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.dop.autotest.testplan.TestPlanService" || ctx.Type() == pb.TestPlanServiceServerType() || ctx.Type() == pb.TestPlanServiceHandlerType():
		return p.TestPlanService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.dop.autotest.testplan", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Description:          "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
