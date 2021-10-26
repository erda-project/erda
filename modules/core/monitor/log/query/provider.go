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

package query

import (
	"time"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/httpserver"
	pb "github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/modules/core/monitor/log/storage"
	"github.com/erda-project/erda/modules/monitor/common"
	monitorperm "github.com/erda-project/erda/modules/monitor/common/permission"
	"github.com/erda-project/erda/pkg/common/apis"
	perm "github.com/erda-project/erda/pkg/common/permission"
)

type config struct{}

type provider struct {
	Cfg                 *config
	Log                 logs.Logger
	Register            transport.Register `autowired:"service-register" optional:"true"`
	Router              httpserver.Router  `autowired:"http-router"`
	Perm                perm.Interface     `autowired:"permission"`
	StorageReader       storage.Storage    `autowired:"log-storage-elasticsearch-reader" optional:"true"`
	K8sReader           storage.Storage    `autowired:"log-storage-kubernetes-reader" optional:"true"`
	FrozenStorageReader storage.Storage    `autowired:"log-storage-cassandra-reader" optional:"true"`

	logQueryService *logQueryService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.logQueryService = &logQueryService{
		p:                   p,
		startTime:           time.Now().UnixNano(),
		storageReader:       p.StorageReader,
		k8sReader:           p.K8sReader,
		frozenStorageReader: p.FrozenStorageReader,
	}
	if p.Register != nil {
		pb.RegisterLogQueryServiceImp(p.Register, p.logQueryService, apis.Options(), p.Perm.Check(
			perm.NoPermMethod(pb.LogQueryServiceServer.GetLog),
			perm.Method(pb.LogQueryServiceServer.GetLogByRuntime, perm.ScopeApp, common.ResourceRuntime, perm.ActionGet, perm.FieldValue("ApplicationId")),
			perm.Method(pb.LogQueryServiceServer.GetLogByOrganization, perm.ScopeOrg, common.ResourceOrgCenter, perm.ActionGet, monitorperm.OrgIDByClusterWrapper("ClusterName")),
		))
	}

	p.initRoutes(p.Router)
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.monitor.log.query.LogQueryService" || ctx.Type() == pb.LogQueryServiceServerType() || ctx.Type() == pb.LogQueryServiceHandlerType():
		return p.logQueryService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.monitor.log.query", &servicehub.Spec{
		Services:   pb.ServiceNames(),
		Types:      pb.Types(),
		ConfigFunc: func() interface{} { return &config{} },
		Creator:    func() servicehub.Provider { return &provider{} },
	})
}
