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

package cron

import (
	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/mysqlxorm"
	pb "github.com/erda-project/erda-proto-go/core/pipeline/cron/pb"
	"github.com/erda-project/erda/modules/pipeline/providers/cron/db"
	"github.com/erda-project/erda/modules/pipeline/services/crondsvc"
)

type config struct {
}

// +provider
type provider struct {
	Cfg         *config
	Log         logs.Logger
	Register    transport.Register
	MySQL       mysqlxorm.Interface `autowired:"mysql-xorm"`
	cronService *Service

	crondSvc *crondsvc.CrondSvc
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.cronService = &Service{
		dbClient: &db.Client{Interface: p.MySQL},
		p:        p,
	}
	if p.Register != nil {
		pb.RegisterCronServiceImp(p.Register, p.cronService)
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.pipeline.cron.CronService" || ctx.Type() == pb.CronServiceServerType() || ctx.Type() == pb.CronServiceHandlerType():
		return p.cronService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.pipeline.cron", &servicehub.Spec{
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
