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

package runtime

import (
	"context"
	"github.com/erda-project/erda/apistructs"
	perm "github.com/erda-project/erda/pkg/common/permission"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	dicehubpb "github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	pipelinepb "github.com/erda-project/erda-proto-go/core/pipeline/pipeline/pb"
	tenantpb "github.com/erda-project/erda-proto-go/msp/tenant/pb"
	"github.com/erda-project/erda-proto-go/orchestrator/runtime/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/pkg/audit"
	"github.com/erda-project/erda/internal/tools/orchestrator/components/addon/mysql"
	"github.com/erda-project/erda/internal/tools/orchestrator/conf"
	"github.com/erda-project/erda/internal/tools/orchestrator/dbclient"
	"github.com/erda-project/erda/internal/tools/orchestrator/events"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/instanceinfo"
	"github.com/erda-project/erda/internal/tools/orchestrator/scheduler/impl/servicegroup"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/addon"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/resource"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/crypto/encryption"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type config struct {
}

// +provider
type provider struct {
	Cfg               *config
	Logger            logs.Logger
	Register          transport.Register
	DB                *gorm.DB                       `autowired:"mysql-client"`
	EventManager      *events.EventManager           `autowired:"erda.orchestrator.events.event-manager"`
	ClusterSvc        clusterpb.ClusterServiceServer `autowired:"erda.core.clustermanager.cluster.ClusterService"`
	RuntimeService    *RuntimeService
	DicehubReleaseSvc dicehubpb.ReleaseServiceServer `autowired:"erda.core.dicehub.release.ReleaseService"`
	Org               org.ClientInterface
	TenantSvc         tenantpb.TenantServiceServer     `autowired:"erda.msp.tenant.TenantService"`
	PipelineSvc       pipelinepb.PipelineServiceServer `autowired:"erda.core.pipeline.pipeline.PipelineService"`
	audit             audit.Auditor
	Bundle            *bundle.Bundle
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.audit = audit.GetAuditor(ctx)
	p.Bundle = bundle.New(
		bundle.WithErdaServer(),
		bundle.WithClusterManager(),
		bundle.WithScheduler(),
	)
	db := NewDBService(p.DB)
	dbClient := &dbclient.DBClient{
		DBEngine: &dbengine.DBEngine{
			DB: p.DB,
		},
	}

	instanceinfoImpl := instanceinfo.NewInstanceInfoImpl()
	scheduler := scheduler.NewScheduler(instanceinfoImpl, p.ClusterSvc)

	encrypt := encryption.New(
		encryption.WithRSAScrypt(encryption.NewRSAScrypt(encryption.RSASecret{
			PublicKey:          conf.PublicKey(),
			PublicKeyDataType:  encryption.Base64,
			PrivateKey:         conf.PrivateKey(),
			PrivateKeyType:     encryption.PKCS1,
			PrivateKeyDataType: encryption.Base64,
		})))

	resource := resource.New(
		resource.WithDBClient(dbClient),
		resource.WithBundle(p.Bundle),
	)

	// init addon
	a := addon.New(
		addon.WithDBClient(dbClient),
		addon.WithBundle(p.Bundle),
		addon.WithResource(resource),
		addon.WithEnvEncrypt(encrypt),
		addon.WithKMSWrapper(mysql.NewKMSWrapper(p.Bundle)),
		addon.WithHTTPClient(httpclient.New(
			httpclient.WithTimeout(time.Second, time.Second*60),
		)),
		addon.WithCap(scheduler.Httpendpoints.Cap),
		addon.WithServiceGroup(scheduler.Httpendpoints.ServiceGroupImpl),
		addon.WithInstanceinfoImpl(instanceinfoImpl),
		addon.WithClusterInfoImpl(scheduler.Httpendpoints.ClusterinfoImpl),
		addon.WithClusterSvc(p.ClusterSvc),
		addon.WithTenantSvc(p.TenantSvc),
		addon.WithOrg(p.Org),
	)

	p.RuntimeService = NewRuntimeService(
		WithBundleService(p.Bundle),
		WithDBService(db),
		WithEventManagerService(p.EventManager),
		WithServiceGroupImpl(servicegroup.NewServiceGroupImplInit()),
		WithClusterSvc(p.ClusterSvc),
		WithReleaseSvc(p.DicehubReleaseSvc),
		WithOrg(p.Org),
		WithClusterInfoImpl(scheduler.Httpendpoints.ClusterinfoImpl),
		WithScheduler(scheduler),
		WithAddon(a),
		WithPipelineSvc(p.PipelineSvc),
	)

	if p.Register != nil {
		type RuntimeServiceHandle = *RuntimeService
		pb.RegisterRuntimePrimaryServiceImp(p.Register, p.RuntimeService, apis.Options(), p.CheckRuntimeID(
			p.Method(RuntimeServiceHandle.GetRuntime, perm.ScopeApp, "runtime-dev", perm.ActionGet, p.GetAppIDByRuntimeID("NameOrID")),
			p.Method(RuntimeServiceHandle.DelRuntime, perm.ScopeApp, "runtime-dev", perm.ActionDelete, p.GetAppIDByRuntimeID("Id")),
			p.Method(RuntimeServiceHandle.StopRuntime, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.ReDeployRuntimeAction, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.ReDeployRuntime, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.RollBackRuntimeAction, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.RollBackRuntime, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),

			p.Method(RuntimeServiceHandle.GetRuntime, perm.ScopeApp, "runtime-test", perm.ActionGet, p.GetAppIDByRuntimeID("NameOrID")),
			p.Method(RuntimeServiceHandle.DelRuntime, perm.ScopeApp, "runtime-test", perm.ActionDelete, p.GetAppIDByRuntimeID("Id")),
			p.Method(RuntimeServiceHandle.StopRuntime, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.ReDeployRuntimeAction, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.ReDeployRuntime, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.RollBackRuntimeAction, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.RollBackRuntime, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),

			p.Method(RuntimeServiceHandle.GetRuntime, perm.ScopeApp, "runtime-staging", perm.ActionGet, p.GetAppIDByRuntimeID("NameOrID")),
			p.Method(RuntimeServiceHandle.DelRuntime, perm.ScopeApp, "runtime-staging", perm.ActionDelete, p.GetAppIDByRuntimeID("Id")),
			p.Method(RuntimeServiceHandle.StopRuntime, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.ReDeployRuntimeAction, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.ReDeployRuntime, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.RollBackRuntimeAction, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.RollBackRuntime, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),

			p.Method(RuntimeServiceHandle.GetRuntime, perm.ScopeApp, "runtime-prod", perm.ActionGet, p.GetAppIDByRuntimeID("NameOrID")),
			p.Method(RuntimeServiceHandle.DelRuntime, perm.ScopeApp, "runtime-prod", perm.ActionDelete, p.GetAppIDByRuntimeID("Id")),
			p.Method(RuntimeServiceHandle.CreateRuntimeByRelease, perm.ScopeApp, "runtime-prod", perm.ActionCreate, p.GetAppIDByRuntimeID("ApplicationId")),
			p.Method(RuntimeServiceHandle.CreateRuntimeByReleaseAction, perm.ScopeApp, "runtime-prod", perm.ActionCreate, p.GetAppIDByRuntimeID("ApplicationId")),
			p.Method(RuntimeServiceHandle.StopRuntime, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.ReDeployRuntimeAction, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.ReDeployRuntime, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.RollBackRuntimeAction, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
			p.Method(RuntimeServiceHandle.RollBackRuntime, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.GetAppIDByRuntimeID("RuntimeID")),
		),
			p.audit.Audit(
				audit.Method(RuntimeServiceHandle.DelRuntime, audit.AppScope, string(apistructs.DeleteRuntimeTemplate),
					func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
						r := resp.(*pb.Runtime)
						return r.ApplicationID, map[string]interface{}{
							"applicationName": r.ApplicationName,
							"workspace":       r.Workspace,
							"runtimeName":     r.Name,
						}, nil
					},
				),

				audit.Method(RuntimeServiceHandle.ReDeployRuntime, audit.AppScope, string(apistructs.RedeployRuntimeTemplate),
					func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
						r := resp.(*pb.DeploymentCreateResponse)
						return r.ApplicationId, map[string]interface{}{}, nil
					},
				),

				audit.Method(RuntimeServiceHandle.ReDeployRuntimeAction, audit.AppScope, string(apistructs.RedeployRuntimeTemplate),
					func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
						r := resp.(*pb.DeploymentCreateResponse)
						return r.ApplicationId, map[string]interface{}{}, nil
					},
				),

				audit.Method(RuntimeServiceHandle.RollBackRuntime, audit.AppScope, string(apistructs.RollbackRuntimeTemplate),
					func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
						r := resp.(*pb.DeploymentCreateResponse)
						return r.ApplicationId, map[string]interface{}{}, nil
					},
				),

				audit.Method(RuntimeServiceHandle.RollBackRuntimeAction, audit.AppScope, string(apistructs.RollbackRuntimeTemplate),
					func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
						r := resp.(*pb.DeploymentCreateResponse)
						return r.ApplicationId, map[string]interface{}{}, nil
					},
				),

				audit.Method(RuntimeServiceHandle.CreateRuntime, audit.AppScope, string(apistructs.DeployRuntimeTemplate),
					func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
						r := resp.(*pb.DeploymentCreateResponse)
						return r.ApplicationId, map[string]interface{}{}, nil
					},
				),

				audit.Method(RuntimeServiceHandle.CreateRuntimeByRelease, audit.AppScope, string(apistructs.DeployRuntimeTemplate),
					func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
						r := resp.(*pb.DeploymentCreateResponse)
						return r.ApplicationId, map[string]interface{}{}, nil
					},
				),

				audit.Method(RuntimeServiceHandle.CreateRuntimeByReleaseAction, audit.AppScope, string(apistructs.DeployRuntimeTemplate),
					func(ctx context.Context, req, resp interface{}, err error) (interface{}, map[string]interface{}, error) {
						r := resp.(*pb.DeploymentCreateResponse)
						return r.ApplicationId, map[string]interface{}{}, nil
					},
				),
			),
		)

		pb.RegisterRuntimeSecondaryServiceImp(p.Register, p.RuntimeService, apis.Options(), p.CheckRuntimeIDs(
			p.Method(RuntimeServiceHandle.ListMyRuntimes, perm.ScopeApp, "runtime-dev", perm.ActionGet, p.FieldValue("AppID")),
			p.Method(RuntimeServiceHandle.ListRuntimesGroupByApps, perm.ScopeApp, "runtime-dev", perm.ActionGet, p.FieldValue("ApplicationID"), func(permission *Permission) {
				permission.skipPermInternalClient = true
			}),

			p.Method(RuntimeServiceHandle.ListMyRuntimes, perm.ScopeApp, "runtime-test", perm.ActionGet, p.FieldValue("AppID")),
			p.Method(RuntimeServiceHandle.ListRuntimesGroupByApps, perm.ScopeApp, "runtime-test", perm.ActionGet, p.FieldValue("ApplicationID"), func(permission *Permission) {
				permission.skipPermInternalClient = true
			}),

			p.Method(RuntimeServiceHandle.ListMyRuntimes, perm.ScopeApp, "runtime-staging", perm.ActionGet, p.FieldValue("AppID")),
			p.Method(RuntimeServiceHandle.ListRuntimesGroupByApps, perm.ScopeApp, "runtime-staging", perm.ActionGet, p.FieldValue("ApplicationID"), func(permission *Permission) {
				permission.skipPermInternalClient = true
			}),

			p.Method(RuntimeServiceHandle.ListMyRuntimes, perm.ScopeApp, "runtime-prod", perm.ActionGet, p.FieldValue("AppID")),
			p.Method(RuntimeServiceHandle.ListRuntimesGroupByApps, perm.ScopeApp, "runtime-prod", perm.ActionGet, p.FieldValue("ApplicationID"), func(permission *Permission) {
				permission.skipPermInternalClient = true
			}),
		))

		pb.RegisterRuntimeTertiaryServiceImp(p.Register, p.RuntimeService, apis.Options(), p.Check(
			p.Method(RuntimeServiceHandle.ListRuntimes, perm.ScopeApp, "runtime-dev", perm.ActionGet, p.FieldValue("ApplicationID")),
			p.Method(RuntimeServiceHandle.CreateRuntime, perm.ScopeApp, "runtime-dev", perm.ActionCreate, p.FieldValue("Extra.ApplicationId")),
			p.Method(RuntimeServiceHandle.CreateRuntimeByRelease, perm.ScopeApp, "runtime-dev", perm.ActionCreate, p.FieldValue("ApplicationId")),
			p.Method(RuntimeServiceHandle.CreateRuntimeByReleaseAction, perm.ScopeApp, "runtime-dev", perm.ActionCreate, p.FieldValue("ApplicationId")),

			p.Method(RuntimeServiceHandle.ListRuntimes, perm.ScopeApp, "runtime-test", perm.ActionGet, p.FieldValue("ApplicationID")),
			p.Method(RuntimeServiceHandle.CreateRuntime, perm.ScopeApp, "runtime-test", perm.ActionCreate, p.FieldValue("Extra.ApplicationId")),
			p.Method(RuntimeServiceHandle.CreateRuntimeByRelease, perm.ScopeApp, "runtime-test", perm.ActionCreate, p.FieldValue("ApplicationId")),
			p.Method(RuntimeServiceHandle.CreateRuntimeByReleaseAction, perm.ScopeApp, "runtime-test", perm.ActionCreate, p.FieldValue("ApplicationId")),

			p.Method(RuntimeServiceHandle.ListRuntimes, perm.ScopeApp, "runtime-staging", perm.ActionGet, p.FieldValue("ApplicationID")),
			p.Method(RuntimeServiceHandle.CreateRuntime, perm.ScopeApp, "runtime-staging", perm.ActionCreate, p.FieldValue("Extra.ApplicationId")),
			p.Method(RuntimeServiceHandle.CreateRuntimeByRelease, perm.ScopeApp, "runtime-staging", perm.ActionCreate, p.FieldValue("ApplicationId")),
			p.Method(RuntimeServiceHandle.CreateRuntimeByReleaseAction, perm.ScopeApp, "runtime-staging", perm.ActionCreate, p.FieldValue("ApplicationId")),

			p.Method(RuntimeServiceHandle.ListRuntimes, perm.ScopeApp, "runtime-prod", perm.ActionGet, p.FieldValue("ApplicationID")),
			p.Method(RuntimeServiceHandle.CreateRuntime, perm.ScopeApp, "runtime-prod", perm.ActionCreate, p.FieldValue("Extra.ApplicationId")),
			p.Method(RuntimeServiceHandle.CreateRuntimeByRelease, perm.ScopeApp, "runtime-prod", perm.ActionCreate, p.FieldValue("ApplicationId")),
			p.Method(RuntimeServiceHandle.CreateRuntimeByReleaseAction, perm.ScopeApp, "runtime-prod", perm.ActionCreate, p.FieldValue("ApplicationId")),

			p.Method(RuntimeServiceHandle.BatchRuntimeService, perm.ScopeApp, "", perm.ActionGet, p.FieldValue(""), func(permission *Permission) {
				permission.skipPermInternalClient = true
			}),
			p.Method(RuntimeServiceHandle.FullGC, perm.ScopeApp, "", perm.ActionGet, p.FieldValue(""), func(permission *Permission) {
				permission.skipPermInternalClient = true
			}),
		))
	}
	return nil
}

func init() {
	servicehub.Register("erda.orchestrator.runtime", &servicehub.Spec{
		Services: append(pb.ServiceNames()),
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
