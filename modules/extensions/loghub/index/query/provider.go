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

package query

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/olivere/elastic"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/extensions/loghub/index/query/db"
	"github.com/erda-project/erda/pkg/httpclient"
)

type define struct{}

func (d *define) Service() []string { return []string{"logs-index-query"} }
func (d *define) Dependencies() []string {
	return []string{"elasticsearch", "elasticsearch@logs", "mysql", "i18n", "http-server"}
}
func (d *define) Summary() string     { return "logs query" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Timeout     time.Duration `file:"timeout" default:"60s"`
	QueryBackES bool          `file:"query_back_es" default:"false"`
}

type provider struct {
	C          *config
	L          logs.Logger
	mysql      *gorm.DB
	client     *elastic.Client
	backClient *elastic.Client // 为了es迁移能够查询到之前es的数据
	db         *db.DB
	bdl        *bundle.Bundle
	t          i18n.Translator
}

func (p *provider) Init(ctx servicehub.Context) error {
	hc := httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	p.bdl = bundle.New(
		bundle.WithHTTPClient(hc),
		bundle.WithCMDB(),
	)
	p.mysql = ctx.Service("mysql").(mysql.Interface).DB()
	p.db = db.New(p.mysql)

	es := ctx.Service("elasticsearch@logs").(elasticsearch.Interface)
	p.client = es.Client()
	backES := ctx.Service("elasticsearch").(elasticsearch.Interface)
	p.backClient = backES.Client()
	if es.URL() == backES.URL() {
		p.C.QueryBackES = false
	}

	p.t = ctx.Service("i18n").(i18n.I18n).Translator("log-metrics")
	routes := ctx.Service("http-server", interceptors.Recover(p.L)).(httpserver.Router)
	return p.intRoutes(routes)
}

// func (p *provider) Start() error { return nil }
// func (p *provider) Close() error { return nil }

func init() {
	servicehub.RegisterProvider("logs-index-query", &define{})
}
