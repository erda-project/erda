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

package settings

import (
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/jinzhu/gorm"
)

type define struct{}

func (d *define) Service() []string      { return []string{"global-settings"} }
func (d *define) Dependencies() []string { return []string{"http-server", "mysql", "i18n"} }
func (d *define) Summary() string        { return "global settings" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type provider struct {
	L      logs.Logger
	db     *gorm.DB
	cfgMap map[string]map[string]*configDefine
	t      i18n.Translator
	bundle *bundle.Bundle
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.t = ctx.Service("i18n").(i18n.I18n).Translator("settings")
	p.initConfigMap()

	p.bundle = bundle.New(
		bundle.WithHTTPClient(httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))),
		bundle.WithCMDB(),
	)

	p.db = ctx.Service("mysql").(mysql.Interface).DB()
	routes := ctx.Service("http-server",
		//telemetry.HttpMetric(),
		interceptors.Recover(p.L)).(httpserver.Router)
	return p.intRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("global-settings", &define{})
}
