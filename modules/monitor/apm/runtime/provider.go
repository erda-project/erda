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

package runtime

import (
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/modules/monitor/common/db"
	"github.com/olivere/elastic"
)

type provider struct {
	Cfg *config
	Log logs.Logger
	db  *db.DB
	es  *elastic.Client
	ctx servicehub.Context
}

type define struct{}

func (d *define) Services() []string     { return []string{"apm-runtime"} }
func (d *define) Dependencies() []string { return []string{"http-server", "elasticsearch", "mysql"} }
func (d *define) Summary() string        { return "apm-runtime api" }
func (d *define) Description() string    { return d.Summary() }
func (d *define) Config() interface{}    { return &config{} }

func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct{}

func (runtime *provider) Init(ctx servicehub.Context) (err error) {
	runtime.ctx = ctx

	// elasticsearch
	es := ctx.Service("elasticsearch").(elasticsearch.Interface)
	runtime.es = es.Client()

	// mysql
	runtime.db = db.New(ctx.Service("mysql").(mysql.Interface).DB())

	routes := ctx.Service("http-server", interceptors.Recover(runtime.Log)).(httpserver.Router)
	return runtime.initRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("apm-runtime", &define{})
}
