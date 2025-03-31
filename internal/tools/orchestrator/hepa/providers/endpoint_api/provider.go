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
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/hepa/endpoint_api/pb"
	projPb "github.com/erda-project/erda-proto-go/core/project/pb"
	runtimePb "github.com/erda-project/erda-proto-go/orchestrator/runtime/pb"
	"github.com/erda-project/erda/internal/pkg/cron"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/common"
	repositoryService "github.com/erda-project/erda/internal/tools/orchestrator/hepa/repository/service"
	"github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/endpoint_api/impl"
	zoneI "github.com/erda-project/erda/internal/tools/orchestrator/hepa/services/zone/impl"
	"github.com/erda-project/erda/pkg/common/apis"
	perm "github.com/erda-project/erda/pkg/common/permission"
)

type config struct {
	ClearEndpointsCronExpr string          `json:"clearEndpointsCronExpr" yaml:"clearEndpointsCronExpr"`
	ClearEndpointsClusters map[string]bool `json:"clearEndpointsClusters" yaml:"clearEndpointsClusters"`
}

// +provider
type provider struct {
	Cfg                *config
	Log                logs.Logger
	Register           transport.Register
	endpointApiService *endpointApiService
	Perm               perm.Interface                         `autowired:"permission"`
	ProjCli            projPb.ProjectServer                   `autowired:"erda.core.project.Project"`
	RuntimeCli         runtimePb.RuntimeTertiaryServiceServer `autowired:"erda.orchestrator.runtime.RuntimeTertiaryService"`
	Cron               cron.Interface                         `autowired:"easy-cron-client"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	runtimeService, err := repositoryService.NewGatewayRuntimeServiceServiceImpl()
	if err != nil {
		return errors.Wrap(err, "failed to NewGatewayRuntimeServiceServiceImpl")
	}
	gatewayRouteService, err := repositoryService.NewGatewayRouteServiceImpl()
	if err != nil {
		return errors.Wrap(err, "failed to NewGatewayRouteServiceImpl")
	}
	gatewayServiceService, err := repositoryService.NewGatewayServiceServiceImpl()
	if err != nil {
		return errors.Wrap(err, "failed to NewGatewayServiceServiceImpl")
	}
	kongInfoService, err := repositoryService.NewGatewayKongInfoServiceImpl()
	if err != nil {
		return errors.Wrap(err, "failed to NewGatewayKongInfoServiceImpl")
	}
	if p.ProjCli == nil {
		p.Log.Fatal("projCli is nil")
	}
	if p.RuntimeCli == nil {
		p.Log.Fatal("runtimeCli is nil")
	}
	p.endpointApiService = &endpointApiService{
		projCli:               p.ProjCli,
		runtimeCli:            p.RuntimeCli,
		runtimeService:        runtimeService,
		gatewayRouteService:   gatewayRouteService,
		gatewayServiceService: gatewayServiceService,
		kongInfoService:       kongInfoService,
		cron:                  p.Cron,
	}
	if clearEndpointsCronExpr := os.Getenv("CLEAR_ENDPOINTS_CRON_EXPR"); clearEndpointsCronExpr != "" {
		p.Cfg.ClearEndpointsCronExpr = clearEndpointsCronExpr
	}
	if clearEndpointsClusters := os.Getenv("CLEAR_ENDPOINTS_CLUSTERS"); clearEndpointsClusters != "" {
		p.Cfg.ClearEndpointsClusters = make(map[string]bool)
		clustersNames := strings.Split(clearEndpointsClusters, ",")
		for _, clusterName := range clustersNames {
			clusterName = strings.TrimLeft(clusterName, " ")
			clusterName = strings.TrimRight(clusterName, " ")
			p.Cfg.ClearEndpointsClusters[clusterName] = true
		}
	}
	if len(p.Cfg.ClearEndpointsClusters) > 0 {
		if _, err := p.Cron.AddFunc(p.Cfg.ClearEndpointsCronExpr, "ClearInvalidEndpointApi", func() bool {
			for clusterName, ok := range p.Cfg.ClearEndpointsClusters {
				if ok {
					if _, err := p.endpointApiService.
						ClearInvalidEndpointApi(ctx, &pb.ListInvalidEndpointApiReq{ClusterName: clusterName}); err != nil {
						p.Log.Warnf("failed to ClearInvalidEndpointApi in the task: %v\n", err)
					}
				}
			}
			return false
		}); err != nil {
			p.Log.Fatal("failed to AddFunc to ClearInvalidEndpointApi")
		}
	}
	err = zoneI.NewGatewayZoneServiceImpl()
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
			perm.Method(apiService.ListInvalidEndpointApi, perm.ScopeOrg, "org", perm.ActionGet, perm.OrgIDValue()),
			perm.Method(apiService.ClearInvalidEndpointApi, perm.ScopeOrg, "org", perm.ActionGet, perm.OrgIDValue()),
			perm.Method(apiService.ListAllCrontabs, perm.ScopeOrg, "org", perm.ActionGet, perm.OrgIDValue()),
			perm.NoPermMethod(apiService.ListPackageApis),
			perm.NoPermMethod(apiService.DeletePackageApis),
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
			"erda.core.project.Project",
			"erda.orchestrator.runtime.RuntimeService",
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
