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

package apis

import (
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/pkg/bundle-ex/cmdb"
	"github.com/erda-project/erda/pkg/httpclient"
)

type define struct{}

func (d *define) Services() []string { return []string{"report-apis"} }
func (d *define) Dependencies() []string {
	return []string{"http-server", "mysql"}
}
func (d *define) Summary() string     { return "report apis" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &provider{} }
}

type config struct {
	Pipeline struct {
		Version       string `file:"version" default:"1.1"`
		ActionType    string `file:"action_type" default:"reportengine"`
		ActionVersion string `file:"action_version" default:"1.0"`
	} `file:"pipeline"`
	ReportCron struct {
		DailyCron   string `file:"daily_cron"`
		WeeklyCron  string `file:"weekly_cron"`
		MonthlyCron string `file:"monthly_cron"`
	} `file:"report_cron"`
	ClusterName  string `env:"DICE_CLUSTER_NAME" default:""`
	DiceProtocol string `env:"DICE_PROTOCOL" default:"http"`
}

type provider struct {
	Cfg  *config
	Log  logs.Logger
	bdl  *bundle.Bundle
	cmdb *cmdb.Cmdb
	t    i18n.Translator
	db   *DB
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.t = ctx.Service("i18n").(i18n.I18n).Translator("alert")
	p.db = newDB(ctx.Service("mysql").(mysql.Interface).DB())
	hc := httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	p.cmdb = cmdb.New(cmdb.WithHTTPClient(hc))
	bundleOpts := []bundle.Option{
		bundle.WithHTTPClient(hc),
		bundle.WithPipeline(),
		bundle.WithCMDB(),
	}

	p.bdl = bundle.New(bundleOpts...)
	routes := ctx.Service("http-server", interceptors.Recover(p.Log)).(httpserver.Router)
	return p.intRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("report-apis", &define{})
}
