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

package horizontalpodscaler

import (
	"context"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/orchestrator/horizontalpodscaler/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/audit"
	"github.com/erda-project/erda/internal/tools/orchestrator/events"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/servicegroup"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

// +provider
type provider struct {
	Cfg             *config
	Log             logs.Logger
	Register        transport.Register
	hpscalerService pb.HPScalerServiceServer //`autowired:"erda.orchestrator.horizontalpodscaler.HPScalerService"`
	audit           audit.Auditor

	DB           *gorm.DB             `autowired:"mysql-client"`
	EventManager *events.EventManager `autowired:"erda.orchestrator.events.event-manager"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.audit = audit.GetAuditor(ctx)
	p.hpscalerService = NewRuntimeHPScalerService(
		WithBundleService(NewBundleService()),
		WithDBService(NewDBService(p.DB)),
		WithServiceGroupImpl(servicegroup.NewServiceGroupImplInit()),
	)

	if p.Register != nil {
		type HPScalerService = pb.HPScalerServiceServer
		pb.RegisterHPScalerServiceImp(p.Register, p.hpscalerService, apis.Options(), p.audit.Audit(
			audit.Method(HPScalerService.CreateRuntimeHPARules, audit.AppScope, string(apistructs.CreateAndApplyHPARule),
				func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
					r := req.(*pb.HPARuleCreateRequest)
					services := make([]string, 0)
					for _, svc := range r.Services {
						services = append(services, svc.ServiceName)
					}
					return services, map[string]interface{}{
						"runtimeId": r.RuntimeID,
					}, nil
				}),
			audit.Method(HPScalerService.UpdateRuntimeHPARules, audit.AppScope, string(apistructs.UpdateHPARule),
				func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
					r := req.(*pb.ErdaRuntimeHPARules)
					services := make([]string, 0)
					serviceToRule := make(map[string]interface{})
					serviceToRule["runtimeId"] = r.RuntimeID
					for _, rule := range r.Rules {
						services = append(services, rule.ServiceName)
						serviceToRule[rule.ServiceName] = rule.RuleID
					}
					return services, serviceToRule, nil
					//return nil, map[string]interface{}{}, nil
				}),
			audit.Method(HPScalerService.DeleteHPARulesByIds, audit.AppScope, string(apistructs.DeleteHPARule),
				func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
					r := req.(*pb.DeleteRuntimeHPARulesRequest)
					rules := make([]string, 0)
					for _, ruleId := range r.Rules {
						rules = append(rules, ruleId)
					}
					return rules, map[string]interface{}{
						"runtimeId": r.RuntimeID,
					}, nil
				}),
			audit.Method(HPScalerService.ApplyOrCancelHPARulesByIds, audit.AppScope, string(apistructs.ApplyOrCancelHPARule),
				func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
					r := req.(*pb.ApplyOrCancelHPARulesRequest)
					rules := make([]string, 0)
					serviceToRule := make(map[string]interface{})
					serviceToRule["runtimeId"] = r.RuntimeID
					for _, ruleAction := range r.RuleAction {
						rules = append(rules, ruleAction.RuleId)
						serviceToRule[ruleAction.RuleId] = ruleAction.Action
					}
					return rules, serviceToRule, nil
					//return nil, map[string]interface{}{}, nil
				}),
		))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.orchestrator.horizontalpodscaler.HPScalerService" || ctx.Type() == pb.HPScalerServiceServerType() || ctx.Type() == pb.HPScalerServiceHandlerType():
		return p.hpscalerService
	}
	return p
}

func init() {
	servicehub.Register("erda.orchestrator.horizontalpodscaler", &servicehub.Spec{
		Services: pb.ServiceNames(),
		Types:    pb.Types(),
		OptionalDependencies: []string{
			"erda.orchestrator.events",
			"service-register",
			"mysql",
		},
		Description: "",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
