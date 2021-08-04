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

package runtimeapis

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/metricq"
)

type provider struct {
	L       logs.Logger
	metricq metricq.Queryer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.metricq = ctx.Service("metrics-query").(metricq.Queryer)
	routes := ctx.Service("http-server", interceptors.Recover(p.L)).(httpserver.Router)
	return p.intRoutes(routes)
}

func init() {
	servicehub.Register("runtime-apis", &servicehub.Spec{
		Services:     []string{"runtime-apis"},
		Dependencies: []string{"http-server", "metrics-query"},
		Description:  "runtime apis",
		Creator:      func() servicehub.Provider { return &provider{} },
	})
}
