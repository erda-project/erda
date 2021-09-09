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

package endpoint_api

import (
	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	pb "github.com/erda-project/erda-proto-go/core/hepa/endpoint_api/pb"
	"github.com/erda-project/erda/modules/hepa/common"
	"github.com/erda-project/erda/modules/hepa/services/endpoint_api/impl"
	zoneI "github.com/erda-project/erda/modules/hepa/services/zone/impl"
	"github.com/erda-project/erda/pkg/common/apis"
	perm "github.com/erda-project/erda/pkg/common/permission"
)

type config struct {
}

// +provider
type provider struct {
	Cfg                *config
	Log                logs.Logger
	Register           transport.Register
	endpointApiService *endpointApiService
	Perm               perm.Interface `autowired:"permission"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.endpointApiService = &endpointApiService{p}
	err := zoneI.NewGatewayZoneServiceImpl()
	if err != nil {
		return err
	}
	err = impl.NewGatewayOpenapiServiceImpl()
	if err != nil {
		return err
	}
	if p.Register != nil {
		type apiService = pb.EndpointApiServiceServer
		pb.RegisterEndpointApiServiceImp(p.Register, p.endpointApiService, apis.Options(), p.Perm.Check(
			perm.Method(apiService.ChangeEndpointRoot, perm.ScopeOrg, "org", perm.ActionGet, perm.OrgIDValue()),
			perm.Method(apiService.CreateEndpoint, perm.ScopeProject, "project", perm.ActionGet, perm.FieldValue("ProjectId")),
			perm.Method(apiService.CreateEndpointApi, perm.ScopeOrg, "org", perm.ActionGet, perm.OrgIDValue()),
			perm.Method(apiService.DeleteEndpoint, perm.ScopeOrg, "org", perm.ActionGet, perm.OrgIDValue()),
			perm.Method(apiService.DeleteEndpointApi, perm.ScopeOrg, "org", perm.ActionGet, perm.OrgIDValue()),
			perm.Method(apiService.GetEndpoint, perm.ScopeOrg, "org", perm.ActionGet, perm.OrgIDValue()),
			perm.Method(apiService.GetEndpointApis, perm.ScopeOrg, "org", perm.ActionGet, perm.OrgIDValue()),
			perm.Method(apiService.GetEndpoints, perm.ScopeProject, "project", perm.ActionGet, perm.FieldValue("ProjectId")),
			perm.Method(apiService.GetEndpointsName, perm.ScopeProject, "project", perm.ActionGet, perm.FieldValue("ProjectId")),
			perm.Method(apiService.UpdateEndpoint, perm.ScopeOrg, "org", perm.ActionGet, perm.OrgIDValue()),
			perm.Method(apiService.UpdateEndpointApi, perm.ScopeOrg, "org", perm.ActionGet, perm.OrgIDValue()),
		), common.AccessLogWrap(common.AccessLog))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.hepa.endpoint_api.EndpointApiService" || ctx.Type() == pb.EndpointApiServiceServerType() || ctx.Type() == pb.EndpointApiServiceHandlerType():
		return p.endpointApiService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.hepa.endpoint_api", &servicehub.Spec{
		Services:             pb.ServiceNames(),
		Types:                pb.Types(),
		OptionalDependencies: []string{"service-register"},
		Dependencies: []string{
			"hepa",
			"erda.core.hepa.api.ApiService",
			"erda.core.hepa.openapi_rule.OpenapiRuleService",
			"erda.core.hepa.openapi_consumer.OpenapiConsumerService",
			"erda.core.hepa.api_policy.ApiPolicyService",
			"erda.core.hepa.domain.DomainService",
			"erda.core.hepa.global.GlobalService",
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
