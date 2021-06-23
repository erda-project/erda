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

package metric

import (
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
	"github.com/erda-project/erda/pkg/common/apis"
)

type config struct {
}

// +provider
type provider struct {
	Register          transport.Register `autowired:"service-register" optional:"true"`
	Metricq           metricq.Queryer    `autowired:"metrics-query"`
	MetricMetaService *metricMetaService
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.MetricMetaService = &metricMetaService{
		p: p,
	}
	if p.Register != nil {
		pb.RegisterMetricMetaServiceImp(p.Register, p.MetricMetaService, apis.Options())
	}
	return nil
}

func (p *provider) Provide(ctx servicehub.DependencyContext, args ...interface{}) interface{} {
	switch {
	case ctx.Service() == "erda.core.monitor.metric.MetricMetaService" || ctx.Type() == pb.MetricMetaServiceServerType() || ctx.Type() == pb.MetricMetaServiceHandlerType():
		return p.MetricMetaService
	}
	return p
}

func init() {
	servicehub.Register("erda.core.monitor.meta", &servicehub.Spec{
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
