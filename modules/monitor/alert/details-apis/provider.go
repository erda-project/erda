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

package details_apis

import (
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
	"github.com/erda-project/erda/modules/pkg/bundle-ex/cmdb"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type define struct{}

type provider struct {
	L       logs.Logger
	metricq metricq.Queryer
	cmdb    *cmdb.Cmdb
}

func (d *define) Service() []string      { return []string{"alert-details-apis"} }
func (d *define) Dependencies() []string { return []string{"http-server", "metrics-query"} }
func (d *define) Summary() string        { return "details apis for alert" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &provider{} }
}

func (p *provider) Init(ctx servicehub.Context) error {
	hc := httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	p.cmdb = cmdb.New(cmdb.WithHTTPClient(hc))
	p.metricq = ctx.Service("metrics-query").(metricq.Queryer)
	routes := ctx.Service("http-server",
		//telemetry.HttpMetric(),
		interceptors.Recover(p.L)).(httpserver.Router)
	return p.intRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("alert-details-apis", &define{})
}
