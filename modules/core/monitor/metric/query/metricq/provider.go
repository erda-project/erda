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

package metricq

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/i18n"
	indexmanager "github.com/erda-project/erda/modules/core/monitor/metric/index"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/chartmeta"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/metricmeta"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/query"

	// v1
	queryv1 "github.com/erda-project/erda/modules/core/monitor/metric/query/query/v1"
	_ "github.com/erda-project/erda/modules/core/monitor/metric/query/query/v1/formats/chart"   //
	_ "github.com/erda-project/erda/modules/core/monitor/metric/query/query/v1/formats/chartv2" //
	_ "github.com/erda-project/erda/modules/core/monitor/metric/query/query/v1/formats/raw"     //
	_ "github.com/erda-project/erda/modules/core/monitor/metric/query/query/v1/language/json"   //
	_ "github.com/erda-project/erda/modules/core/monitor/metric/query/query/v1/language/params" //
)

type config struct {
	ChartMeta struct {
		Path           string        `file:"path"`
		ReloadInterval time.Duration `file:"reload_interval"`
	} `file:"chart_meta"`
}

type provider struct {
	C          *config
	L          logs.Logger
	Meta       *metricmeta.Manager `autowired:"erda.core.monitor.metric.meta"`
	Index      indexmanager.Index  `autowired:"erda.core.monitor.metric.index"`
	DB         *gorm.DB            `autowired:"mysql-client"`
	ChartTrans i18n.Translator     `autowired:"i18n" translator:"charts"`
	q          *metricq
}

func (p *provider) Init(ctx servicehub.Context) error {
	charts := chartmeta.NewManager(p.DB, p.C.ChartMeta.ReloadInterval, p.C.ChartMeta.Path, p.ChartTrans, p.L)
	err := charts.Init()
	if err != nil {
		return fmt.Errorf("fail to start charts manager: %s", err)
	}

	p.q = &metricq{
		Queryer:   query.New(p.Index),
		queryv1:   queryv1.New(p.Index, charts, p.Meta, p.ChartTrans),
		index:     p.Index,
		meta:      p.Meta,
		charts:    charts,
		handler:   p.queryMetrics,
		handlerV1: p.queryMetricsV1,
	}
	Q = p.q

	routes := ctx.Service("http-server", interceptors.Recover(p.L), interceptors.CORS()).(httpserver.Router)
	err = p.initRoutes(routes)
	if err != nil {
		return fmt.Errorf("fail to init routes: %s", err)
	}
	return nil
}

// Provide .
func (p *provider) Provide(ctx servicehub.DependencyContext, options ...interface{}) interface{} {
	return p.q
}

func init() {
	servicehub.Register("metrics-query-compatibility", &servicehub.Spec{
		Services:     []string{"metrics-query"},
		Dependencies: []string{"http-server"},
		Description:  "metrics query api",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
