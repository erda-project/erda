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

package topology

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/olivere/elastic"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/metricq"
	"github.com/erda-project/erda/modules/monitor/common/db"
)

type provider struct {
	Cfg              *config
	Log              logs.Logger
	db               *db.DB
	es               *elastic.Client
	ctx              servicehub.Context
	metricq          metricq.Queryer
	t                i18n.Translator
	cassandraSession *gocql.Session
}

type define struct{}

func (d *define) Services() []string { return []string{"apm-topology"} }
func (d *define) Dependencies() []string {
	return []string{"http-server", "metrics-query", "elasticsearch", "mysql", "i18n"}
}
func (d *define) Summary() string     { return "apm-topology api" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }

func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	Cassandra cassandra.SessionConfig `file:"cassandra"`
}

func (topology *provider) Init(ctx servicehub.Context) (err error) {
	topology.ctx = ctx

	// elasticsearch
	es := ctx.Service("elasticsearch").(elasticsearch.Interface)
	topology.es = es.Client()

	// translator
	topology.t = ctx.Service("i18n").(i18n.I18n).Translator("topology")

	// mysql
	topology.db = db.New(ctx.Service("mysql").(mysql.Interface).DB())

	topology.metricq = ctx.Service("metrics-query").(metricq.Queryer)

	c := ctx.Service("cassandra").(cassandra.Interface)
	session, err := c.Session(&topology.Cfg.Cassandra)
	topology.cassandraSession = session
	if err != nil {
		return fmt.Errorf("fail to create cassandra session: %s", err)
	}

	routes := ctx.Service("http-server", interceptors.Recover(topology.Log)).(httpserver.Router)
	return topology.initRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("apm-topology", &define{})
}
