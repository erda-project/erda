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

package nodetopo

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
)

type define struct{}

func (d *define) Services() []string     { return []string{"node-topo"} }
func (d *define) Dependencies() []string { return []string{"http-server", "metrics-query", "i18n"} }
func (d *define) Summary() string        { return "node topology" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct{}

type provider struct {
	C       *config
	L       logs.Logger
	metricq metricq.Queryer
	t       i18n.Translator
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.t = ctx.Service("i18n").(i18n.I18n).Translator("node-topo")
	p.metricq = ctx.Service("metrics-query").(metricq.Queryer)
	routes := ctx.Service("http-server", interceptors.Recover(p.L)).(httpserver.Router)
	return p.intRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("node-topo", &define{})
}
