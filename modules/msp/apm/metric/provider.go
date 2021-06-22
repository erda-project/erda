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
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/modules/monitor/common/db"
)

const PATH_PREFIX = "/api/tmc"

type provider struct {
	Log logs.Logger
	db  *db.DB
	Cfg *config
}

type define struct{}

func (d *define) Services() []string { return []string{"erda.msp.apm.metric"} }
func (d *define) Dependencies() []string {
	return []string{"mysql", "i18n"}
}
func (d *define) Summary() string     { return "erda.msp.apm.metric api" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	MonitorAddr                 string `env:"MONITOR_ADDR" default:"monitor.default.svc.cluster.local:7096"`
	MonitorServiceMetricApiPath string `env:"MONITOR_METRIC_PATH" default:"/api/metrics"`
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.db = db.New(ctx.Service("mysql").(mysql.Interface).DB())
	routes := ctx.Service("http-server").(httpserver.Router)
	return p.initRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("erda.msp.apm.metric", &define{})
}


