// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package trace

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-infra/providers/i18n"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda-proto-go/msp/apm/trace/pb"
	"github.com/erda-project/erda/modules/msp/apm/trace/db"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
	Cassandra cassandra.SessionConfig `file:"cassandra"`
}

// +provider
type provider struct {
	Cfg              *config
	Log              logs.Logger
	traceService     *traceService
	I18n             i18n.Translator              `autowired:"i18n" translator:"msp-i18n"`
	Register         transport.Register           `autowired:"service-register"`
	Metric           metricpb.MetricServiceServer `autowired:"erda.core.monitor.metric.MetricService"`
	DB               *gorm.DB                     `autowired:"mysql-client"`
	Cassandra        cassandra.Interface          `autowired:"cassandra"`
	cassandraSession *cassandra.Session
}

func (p *provider) Init(ctx servicehub.Context) error {
	// translator

	session, err := p.Cassandra.NewSession(&p.Cfg.Cassandra)
	if err != nil {
		return fmt.Errorf("fail to create cassandra session: %s", err)
	}
	p.cassandraSession = session
	p.traceService = &traceService{
		p:                     p,
		i18n:                  p.I18n,
		traceRequestHistoryDB: &db.TraceRequestHistoryDB{DB: p.DB},
	}
	if p.Register != nil {
		pb.RegisterTraceServiceImp(p.Register, p.traceService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.msp.apm.trace.TraceService" || ctx.Type() == pb.TraceServiceServerType() || ctx.Type() == pb.TraceServiceHandlerType():
		return p.traceService
	}
	return p
}

func init() {
	servicehub.Register("erda.msp.apm.trace", &servicehub.Spec{
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
