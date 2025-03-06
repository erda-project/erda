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
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

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
	perm "github.com/erda-project/erda/pkg/common/permission"
	"github.com/erda-project/erda/pkg/crypto/encryption"
	"github.com/erda-project/erda/pkg/database/dbengine"
	"github.com/erda-project/erda/pkg/http/httpclient"

	"github.com/jinzhu/gorm"
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
	runtimeService    pb.RuntimeServiceServer
	DicehubReleaseSvc dicehubpb.ReleaseServiceServer `autowired:"erda.core.dicehub.release.ReleaseService"`
	Org               org.ClientInterface
	TenantSvc         tenantpb.TenantServiceServer     `autowired:"erda.msp.tenant.TenantService"`
	Perm              perm.Interface                   `autowired:"permission"`
	PipelineSvc       pipelinepb.PipelineServiceServer `autowired:"erda.core.pipeline.pipeline.PipelineService"`
}

func (p *provider) Init(ctx servicehub.Context) error {
	bdl := bundle.New(
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
		resource.WithBundle(bdl),
	)

	// init addon
	a := addon.New(
		addon.WithDBClient(dbClient),
		addon.WithBundle(bdl),
		addon.WithResource(resource),
		addon.WithEnvEncrypt(encrypt),
		addon.WithKMSWrapper(mysql.NewKMSWrapper(bdl)),
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

	p.runtimeService = NewRuntimeService(
		WithBundleService(bdl),
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
		type RuntimeService = pb.RuntimeServiceHandler
		pb.RegisterRuntimeServiceImp(p.Register, p.runtimeService, apis.Options(),
			p.Perm.Check(
				perm.Method(RuntimeService.GetRuntime, perm.ScopeApp, "runtime-dev", perm.ActionGet, p.FieldValue("AppID")),
				perm.Method(RuntimeService.ListRuntimes, perm.ScopeApp, "runtime-dev", perm.ActionList, p.FieldValue("ApplicationID")),
				perm.Method(RuntimeService.ListRuntimesGroupByApps, perm.ScopeApp, "runtime-dev", perm.ActionList, p.FieldValue("ApplicationID")),
				perm.Method(RuntimeService.ListMyRuntimes, perm.ScopeApp, "runtime-dev", perm.ActionList, p.FieldValue("AppID")),
				perm.Method(RuntimeService.CreateRuntime, perm.ScopeApp, "runtime-dev", perm.ActionCreate, p.FieldValue("Extra.ApplicationId")),
				perm.Method(RuntimeService.DelRuntime, perm.ScopeApp, "runtime-dev", perm.ActionDelete, p.FieldValue("Id")),
				perm.Method(RuntimeService.CreateRuntimeByRelease, perm.ScopeApp, "runtime-dev", perm.ActionCreate, p.FieldValue("ApplicationId")),
				perm.Method(RuntimeService.CreateRuntimeByReleaseAction, perm.ScopeApp, "runtime-dev", perm.ActionCreate, p.FieldValue("ApplicationId")),
				perm.Method(RuntimeService.StopRuntime, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.ReDeployRuntimeAction, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.ReDeployRuntime, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RollBackRuntimeAction, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RollBackRuntime, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RuntimeLogs, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.FieldValue("Id")),
				perm.Method(RuntimeService.CountPRByWorkspace, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.FieldValue("AppID")),
				perm.Method(RuntimeService.StartRuntime, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RestartRuntime, perm.ScopeApp, "runtime-dev", perm.ActionOperate, p.FieldValue("RuntimeID")),

				perm.Method(RuntimeService.GetRuntime, perm.ScopeApp, "runtime-test", perm.ActionGet, p.FieldValue("AppID")),
				perm.Method(RuntimeService.ListRuntimes, perm.ScopeApp, "runtime-test", perm.ActionList, p.FieldValue("ApplicationID")),
				perm.Method(RuntimeService.ListRuntimesGroupByApps, perm.ScopeApp, "runtime-test", perm.ActionList, p.FieldValue("ApplicationID")),
				perm.Method(RuntimeService.ListMyRuntimes, perm.ScopeApp, "runtime-test", perm.ActionList, p.FieldValue("AppID")),
				perm.Method(RuntimeService.CreateRuntime, perm.ScopeApp, "runtime-test", perm.ActionCreate, p.FieldValue("Extra.ApplicationId")),
				perm.Method(RuntimeService.DelRuntime, perm.ScopeApp, "runtime-test", perm.ActionDelete, p.FieldValue("Id")),
				perm.Method(RuntimeService.CreateRuntimeByRelease, perm.ScopeApp, "runtime-test", perm.ActionCreate, p.FieldValue("ApplicationId")),
				perm.Method(RuntimeService.CreateRuntimeByReleaseAction, perm.ScopeApp, "runtime-test", perm.ActionCreate, p.FieldValue("ApplicationId")),
				perm.Method(RuntimeService.StopRuntime, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.ReDeployRuntimeAction, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.ReDeployRuntime, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RollBackRuntimeAction, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RollBackRuntime, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RuntimeLogs, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.FieldValue("Id")),
				perm.Method(RuntimeService.CountPRByWorkspace, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.FieldValue("AppID")),
				perm.Method(RuntimeService.StartRuntime, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RestartRuntime, perm.ScopeApp, "runtime-test", perm.ActionOperate, p.FieldValue("RuntimeID")),

				perm.Method(RuntimeService.GetRuntime, perm.ScopeApp, "runtime-staging", perm.ActionGet, p.FieldValue("AppID")),
				perm.Method(RuntimeService.ListRuntimes, perm.ScopeApp, "runtime-staging", perm.ActionList, p.FieldValue("ApplicationID")),
				perm.Method(RuntimeService.ListRuntimesGroupByApps, perm.ScopeApp, "runtime-staging", perm.ActionList, p.FieldValue("ApplicationID")),
				perm.Method(RuntimeService.ListMyRuntimes, perm.ScopeApp, "runtime-staging", perm.ActionList, p.FieldValue("AppID")),
				perm.Method(RuntimeService.CreateRuntime, perm.ScopeApp, "runtime-staging", perm.ActionCreate, p.FieldValue("Extra.ApplicationId")),
				perm.Method(RuntimeService.DelRuntime, perm.ScopeApp, "runtime-staging", perm.ActionDelete, p.FieldValue("Id")),
				perm.Method(RuntimeService.CreateRuntimeByRelease, perm.ScopeApp, "runtime-staging", perm.ActionCreate, p.FieldValue("ApplicationId")),
				perm.Method(RuntimeService.CreateRuntimeByReleaseAction, perm.ScopeApp, "runtime-staging", perm.ActionCreate, p.FieldValue("ApplicationId")),
				perm.Method(RuntimeService.StopRuntime, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.ReDeployRuntimeAction, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.ReDeployRuntime, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RollBackRuntimeAction, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RollBackRuntime, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RuntimeLogs, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.FieldValue("Id")),
				perm.Method(RuntimeService.CountPRByWorkspace, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.FieldValue("AppID")),
				perm.Method(RuntimeService.StartRuntime, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RestartRuntime, perm.ScopeApp, "runtime-staging", perm.ActionOperate, p.FieldValue("RuntimeID")),

				perm.Method(RuntimeService.GetRuntime, perm.ScopeApp, "runtime-prod", perm.ActionGet, p.FieldValue("AppID")),
				perm.Method(RuntimeService.ListRuntimes, perm.ScopeApp, "runtime-prod", perm.ActionList, p.FieldValue("ApplicationID")),
				perm.Method(RuntimeService.ListRuntimesGroupByApps, perm.ScopeApp, "runtime-prod", perm.ActionList, p.FieldValue("ApplicationID")),
				perm.Method(RuntimeService.ListMyRuntimes, perm.ScopeApp, "runtime-prod", perm.ActionList, p.FieldValue("AppID")),
				perm.Method(RuntimeService.CreateRuntime, perm.ScopeApp, "runtime-prod", perm.ActionCreate, p.FieldValue("Extra.ApplicationId")),
				perm.Method(RuntimeService.DelRuntime, perm.ScopeApp, "runtime-prod", perm.ActionDelete, p.FieldValue("Id")),
				perm.Method(RuntimeService.CreateRuntimeByRelease, perm.ScopeApp, "runtime-prod", perm.ActionCreate, p.FieldValue("ApplicationId")),
				perm.Method(RuntimeService.CreateRuntimeByReleaseAction, perm.ScopeApp, "runtime-prod", perm.ActionCreate, p.FieldValue("ApplicationId")),
				perm.Method(RuntimeService.StopRuntime, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.ReDeployRuntimeAction, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.ReDeployRuntime, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RollBackRuntimeAction, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RollBackRuntime, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RuntimeLogs, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.FieldValue("Id")),
				perm.Method(RuntimeService.CountPRByWorkspace, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.FieldValue("AppID")),
				perm.Method(RuntimeService.StartRuntime, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.FieldValue("RuntimeID")),
				perm.Method(RuntimeService.RestartRuntime, perm.ScopeApp, "runtime-prod", perm.ActionOperate, p.FieldValue("RuntimeID")),
			))
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.orchestrator.runtime.RuntimeService" || ctx.Type() == pb.RuntimeServiceServerType() || ctx.Type() == pb.RuntimeServiceHandlerType():
		return p.runtimeService
	}

	return p
}

// FieldValue 从请求体中获取到appID，或 applicationID
func (p *provider) FieldValue(field string) func(ctx context.Context, req interface{}) (string, error) {
	fields := strings.Split(field, ".")
	last := len(fields) - 1

	result := func(ctx context.Context, req interface{}) (string, error) {
		if value := req; value != nil {
			for i, field := range fields {
				val := reflect.ValueOf(value)
				fmt.Printf("val: %v\n", val)
				for val.Kind() == reflect.Ptr {
					val = val.Elem()
				}
				if val.Kind() != reflect.Struct {
					return "", fmt.Errorf("invalid request type")
				}
				val = val.FieldByName(field)
				if !val.IsValid() {
					break
				}
				value = val.Interface()

				if value == nil {
					break
				}

				// 如果是ID，需要从数据库中获取到appID
				if field == "Id" || field == "RuntimeID" {
					var runtime dbclient.Runtime
					if err := p.DB.Where("id = ?", value).First(&runtime).Error; err != nil {
						if gorm.IsRecordNotFoundError(err) {
							return "", nil
						}
					}
					appId := strconv.FormatUint(runtime.ApplicationID, 10)
					fmt.Printf("appId: %v\n", appId)
					return appId, nil
				}

				if i == last {
					fmt.Printf("value: %v\n", value)
					return fmt.Sprint(value), nil
				}
			}
		}
		return "", fmt.Errorf("not found id for permission")
	}

	return result
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
