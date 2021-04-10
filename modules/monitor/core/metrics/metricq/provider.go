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

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"

	//"github.com/erda-project/erda-infra/providers/httpserver"
	//"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-infra/providers/mysql"
	indexmanager "github.com/erda-project/erda/modules/monitor/core/metrics/index"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/chartmeta"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/metricmeta"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query"
	"github.com/recallsong/go-utils/ioutil"

	_ "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql/formats/chartv2"  //
	_ "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql/formats/dict"     //
	_ "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql/formats/influxdb" //
	_ "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/es-tsql/influxql"         //

	// v1
	queryv1 "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query/v1"
	_ "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query/v1/formats/chart"   //
	_ "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query/v1/formats/chartv2" //
	_ "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query/v1/formats/raw"     //
	_ "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query/v1/language/json"   //
	_ "github.com/erda-project/erda/modules/monitor/core/metrics/metricq/query/v1/language/params" //
)

type define struct{}

func (d *define) Service() []string { return []string{"metrics-query"} }
func (d *define) Dependencies() []string {
	return []string{"mysql", "metrics-index-manager", "http-server", "i18n", "i18n@metric"}
}
func (d *define) Summary() string     { return "metrics query api" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	ChartMeta struct {
		Path           string        `file:"path"`
		ReloadInterval time.Duration `file:"reload_interval"`
	} `file:"chart_meta"`
	MetricMeta struct {
		Sources        []string `file:"sources"`
		GroupFiles     []string `file:"group_files"`
		MetricMetaPath string   `file:"metric_meta_path"`
	} `file:"metric_meta"`
}

type provider struct {
	C *config
	L logs.Logger
	q *metricq
}

func (p *provider) Init(ctx servicehub.Context) error {

	db := ctx.Service("mysql").(mysql.Interface).DB()
	index := ctx.Service("metrics-index-manager").(indexmanager.Index)

	trans := ctx.Service("i18n").(i18n.I18n).Translator("charts")
	charts := chartmeta.NewManager(db, p.C.ChartMeta.ReloadInterval, p.C.ChartMeta.Path, trans, p.L)

	meta := metricmeta.NewManager(
		p.C.MetricMeta.Sources,
		db,
		index,
		p.C.MetricMeta.MetricMetaPath,
		p.C.MetricMeta.GroupFiles,
		ctx.Service("i18n@metric").(i18n.I18n),
		p.L,
	)

	p.q = &metricq{
		Queryer:   query.New(index),
		queryv1:   queryv1.New(index, charts, meta, trans),
		index:     index,
		meta:      meta,
		charts:    charts,
		handler:   p.queryMetrics,
		handlerV1: p.queryMetricsV1,
	}
	Q = p.q

	err := meta.Start()
	if err != nil {
		return fmt.Errorf("fail to start metricmeta manager: %s", err)
	}
	err = charts.Start()
	if err != nil {
		return fmt.Errorf("fail to start charts manager: %s", err)
	}

	// routes := ctx.Service("http-server", metrics.HttpMetric(), interceptors.Recover(p.L), interceptors.CORS()).(httpserver.Router)
	// err = p.initRoutes(routes)
	// if err != nil {
	// 	return fmt.Errorf("fail to init routes: %s", err)
	// }
	return nil
}

// Start .
func (p *provider) Start() error { return nil }
func (p *provider) Close() error {
	return ioutil.CloseMulti(p.q.charts, p.q.meta)
}

// Provide .
func (p *provider) Provide(name string, args ...interface{}) interface{} {
	return p.q
}

func init() {
	servicehub.RegisterProvider("metrics-query", &define{})
}
