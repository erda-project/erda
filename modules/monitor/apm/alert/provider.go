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

package alert

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/mysql"
	apis "github.com/erda-project/erda/modules/monitor/alert/alert-apis"
	"github.com/erda-project/erda/modules/monitor/common/db"
)

type provider struct {
	C             *config
	L             logs.Logger
	authDb        *db.DB
	microAlertAPI apis.MicroAlertAPI
}

type define struct{}

func (d *define) Service() []string      { return []string{"apm-alert"} }
func (d *define) Dependencies() []string { return []string{"http-server", "alert-apis", "mysql"} }
func (d *define) Summary() string        { return "apm-alert api" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }

func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct{}

func (p *provider) Init(ctx servicehub.Context) (err error) {
	p.microAlertAPI = ctx.Service("alert-apis").(apis.MicroAlertAPI)

	// mysql
	p.authDb = db.New(ctx.Service("mysql").(mysql.Interface).DB())

	routes := ctx.Service("http-server", interceptors.Recover(p.L)).(httpserver.Router)
	return p.initRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("apm-alert", &define{})
}
