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
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/modules/monitor/common/db"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
	"github.com/erda-project/erda/modules/monitor/trace/query"
)

type provider struct {
	Cfg     *config
	Log     logs.Logger
	authDb  *db.DB
	metricq metricq.Queryer
	spanq   query.SpanQueryAPI
}

type define struct{}

func (d *define) Services() []string { return []string{"apm-trace"} }
func (d *define) Dependencies() []string {
	return []string{"http-server", "mysql", "metrics-query", "trace-query"}
}
func (d *define) Summary() string     { return "apm-trace api" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }

func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct{}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.metricq = ctx.Service("metrics-query").(metricq.Queryer)

	p.spanq = ctx.Service("trace-query").(query.SpanQueryAPI)
	// mysql
	p.authDb = db.New(ctx.Service("mysql").(mysql.Interface).DB())

	routes := ctx.Service("http-server", interceptors.Recover(p.Log)).(httpserver.Router)
	return p.initRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("apm-trace", &define{})
}
