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

package template

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/modules/monitor/common/db"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
)

type define struct{}

func (d *define) Services() []string { return []string{"template"} }
func (d *define) Dependencies() []string {
	return []string{"http-server", "mysql"}
}
func (d *define) Summary() string     { return "template" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &pconfig{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &provider{} }
}

type pconfig struct {
	PresetDashboards string `file:"preset_dashboards"`
	Tables           struct {
		Template string `file:"template" default:"sp_dashboard_template"`
	} `file:"tables"`
}

type provider struct {
	C         *pconfig
	L         logs.Logger
	metricq   metricq.Queryer
	db        *DB
	authDb    *db.DB
	presetMap map[string][]string
}

// Init .
func (p *provider) Init(ctx servicehub.Context) error {
	if len(p.C.Tables.Template) > 0 {
		tableTemplate = p.C.Tables.Template
	}

	p.authDb = db.New(ctx.Service("mysql").(mysql.Interface).DB())

	p.db = newDB(ctx.Service("mysql").(mysql.Interface).DB())

	routes := ctx.Service("http-server", interceptors.Recover(p.L)).(httpserver.Router)

	return p.intRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("template", &define{})
}
