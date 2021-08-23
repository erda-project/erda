// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	"github.com/erda-project/erda/pkg/http/httpclient"
)

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
		bundle.WithCoreServices(),
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

func init() {
	servicehub.Register("logs-index-query", &servicehub.Spec{
		Services:     []string{"logs-index-query"},
		Dependencies: []string{"elasticsearch", "elasticsearch@logs", "mysql", "i18n", "http-server"},
		Description:  "logs query",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
