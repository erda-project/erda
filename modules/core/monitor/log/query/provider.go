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
	"fmt"
	"time"

	logs "github.com/erda-project/erda-infra/base/logs"
	servicehub "github.com/erda-project/erda-infra/base/servicehub"
	transport "github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-infra/providers/httpserver"
	pb "github.com/erda-project/erda-proto-go/core/monitor/log/query/pb"
	"github.com/erda-project/erda/modules/monitor/common"
	monitorperm "github.com/erda-project/erda/modules/monitor/common/permission"
	"github.com/erda-project/erda/pkg/common/apis"
	perm "github.com/erda-project/erda/pkg/common/permission"
)

type config struct {
	Cassandra cassandra.SessionConfig `file:"cassandra"`
	Download  struct {
		TimeSpan time.Duration `file:"time_span" default:"5m"`
	} `file:"download"`
}

// +provider
type provider struct {
	Cfg             *config
	Logger          logs.Logger
	Register        transport.Register
	logQueryService *logQueryService
	Cassandra       cassandra.Interface
	HttpServer      httpserver.Router `autowired:"http-server"`
	Perm            perm.Interface    `autowired:"permission"`

	cqlQuery CQLQueryInf
}

func (p *provider) Init(ctx servicehub.Context) error {
	session, err := p.Cassandra.Session(&p.Cfg.Cassandra)
	if err != nil {
		return fmt.Errorf("fail to create cassandra session: %s", err)
	}
	p.cqlQuery = &cassandraQuery{
		session: session,
	}

	p.intRoutes(p.HttpServer)

	p.logQueryService = &logQueryService{p}
	if p.Register != nil {
		pb.RegisterLogQueryServiceImp(p.Register, p.logQueryService, apis.Options(), p.Perm.Check(
			perm.NoPermMethod(pb.LogQueryServiceServer.GetLog),
			perm.Method(pb.LogQueryServiceServer.GetLogByRuntime, perm.ScopeApp, common.ResourceRuntime, perm.ActionGet, perm.FieldValue("ApplicationId")),
			perm.Method(pb.LogQueryServiceServer.GetLogByOrganization, perm.ScopeOrg, common.ResourceOrgCenter, perm.ActionGet, monitorperm.OrgIDByClusterWrapper("ClusterName")),
		))
	}
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
