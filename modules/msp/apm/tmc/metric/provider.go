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
	"github.com/erda-project/erda/modules/msp/apm/tmc/tmcconfig"
)

const PATH_PREFIX = "/api/tmc"

type provider struct {
	Log logs.Logger
	db  *db.DB
	cfg tmcconfig.Monitor
}

type define struct{}

func (d *define) Services() []string { return []string{"erda.msp.tmc.metric"} }
func (d *define) Dependencies() []string {
	return []string{"mysql", "i18n"}
}
func (d *define) Summary() string     { return "erda.msp.tmc.metric api" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.cfg = tmcconfig.Conf
	p.db = db.New(ctx.Service("mysql").(mysql.Interface).DB())
	routes := ctx.Service("http-server").(httpserver.Router)
	return p.initRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("erda.msp.tmc.metric", &define{})
}

func (p *provider) initRoutes(routes httpserver.Router) error {
	routes.GET(PATH_PREFIX+"/metrics-query", p.metricQueryByQL)
	routes.POST(PATH_PREFIX+"/metrics-query", p.metricQueryByQL)
	routes.GET(PATH_PREFIX+"/metrics/:scope", p.metricQuery)
	routes.GET(PATH_PREFIX+"/metrics/:scope/histogram", p.metricQueryHistogram)
	routes.GET(PATH_PREFIX+"/metrics/:scope/range", p.metricQueryRange)
	routes.GET(PATH_PREFIX+"/metrics/:scope/apdex", p.metricQueryApdex)
	routes.GET(PATH_PREFIX+"/metric/groups", p.listGroups)
	routes.GET(PATH_PREFIX+"/metric/groups/:id", p.getGroup)
	return nil
}
