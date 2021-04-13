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

package report

import (
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/jinzhu/gorm"
)

type provider struct {
	Cfg    *config
	Log    logs.Logger
	db     *gorm.DB
	ctx    servicehub.Context
	bundle *bundle.Bundle
}

type define struct{}

func (d *define) Services() []string     { return []string{"apm-report"} }
func (d *define) Dependencies() []string { return []string{"http-server", "mysql"} }
func (d *define) Summary() string        { return "apm-report api" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }

func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct{}

func (report *provider) Init(ctx servicehub.Context) (err error) {
	report.ctx = ctx

	// mysql
	report.db = ctx.Service("mysql").(mysql.Interface).DB()

	// bundle
	hc := httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	report.bundle = bundle.New(
		bundle.WithHTTPClient(hc),
		bundle.WithCMDB(),
	)

	// http server
	routes := ctx.Service("http-server", interceptors.Recover(report.Log)).(httpserver.Router)
	return report.initRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("apm-report", &define{})
}
