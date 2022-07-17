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

package rule

import (
	gojsonnet "github.com/google/go-jsonnet"
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	pb "github.com/erda-project/erda-proto-go/core/rule/pb"
	coredao "github.com/erda-project/erda/internal/core/legacy/dao"
	"github.com/erda-project/erda/internal/core/rule/dao"
	"github.com/erda-project/erda/internal/core/rule/executor"
	"github.com/erda-project/erda/internal/core/rule/jsonnet"
)

type config struct {
}

type provider struct {
	Cfg          *config
	Log          logs.Logger
	Register     transport.Register
	DB           *gorm.DB `autowired:"mysql-client"`
	ruleExecutor executor.Executor
	db           *dao.DBClient

	ruleService *ruleService
}

func (p *provider) Init(ctx servicehub.Context) error {
	db := &dao.DBClient{DBClient: &coredao.DBClient{DB: p.DB}}
	p.db = db
	e := &executor.ExprExecutor{
		DB: db,
	}
	p.ruleExecutor = executor.Executor{
		RuleSetExecutor: e,
		TemplateParser: jsonnet.Engine{
			JsonnetVM: gojsonnet.MakeVM(),
		},
	}
	p.ruleService = &ruleService{p}

	if p.Register != nil {
		pb.RegisterRuleServiceImp(p.Register, p.ruleService)
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.rule.RuleService" || ctx.Type() == pb.RuleServiceServerType() || ctx.Type() == pb.RuleServiceHandlerType():
		return p.ruleService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.rule", &servicehub.Spec{
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
