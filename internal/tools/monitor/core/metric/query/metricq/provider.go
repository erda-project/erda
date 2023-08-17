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

package metricq

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric/query/chartmeta"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric/query/metricmeta"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric/query/query"
	queryv1 "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/query/v1"
	_ "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/query/v1/formats/chart"   //
	_ "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/query/v1/formats/chartv2" //
	_ "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/query/v1/formats/raw"     //
	_ "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/query/v1/language/json"   //
	_ "github.com/erda-project/erda/internal/tools/monitor/core/metric/query/query/v1/language/params" //
	"github.com/erda-project/erda/internal/tools/monitor/core/metric/storage"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/meta"
	indexloader "github.com/erda-project/erda/internal/tools/monitor/core/storekit/elasticsearch/index/loader"
)

type config struct {
	ChartMeta struct {
		Path           string        `file:"path"`
		ReloadInterval time.Duration `file:"reload_interval"`
	} `file:"chart_meta"`

	MetricMeta struct {
		MetricMetaCacheExpiration time.Duration `file:"metric_meta_cache_expiration"`
		Sources                   []string      `file:"sources"`
		GroupFiles                []string      `file:"group_files"`
		MetricMetaPath            string        `file:"metric_meta_path"`
	} `file:"metric_meta"`
}

type provider struct {
	C          *config
	L          logs.Logger
	Meta       *metricmeta.Manager   `autowired:"erda.core.monitor.metric.meta"`
	Index      indexloader.Interface `autowired:"elasticsearch.index.loader@metric" optional:"true"`
	DB         *gorm.DB              `autowired:"mysql-client"`
	ChartTrans i18n.Translator       `autowired:"i18n" translator:"charts"`
	q          *Metricq

	Log        logs.Logger
	Register   transport.Register `autowired:"service-register" optional:"true"`
	MetricTran i18n.I18n          `autowired:"i18n@metric"`
	Redis      *redis.Client      `autowired:"redis-client"`

	Storage storage.Storage `autowired:"metric-storage" optional:"true"`

	CkMetaLoader    meta.Interface  `autowired:"clickhouse.meta.loader@metric" optional:"true"`
	CkStorageReader storage.Storage `autowired:"metric-storage-clickhouse" optional:"true"`

	Org org.ClientInterface
}

func (p *provider) Init(ctx servicehub.Context) error {
	charts := chartmeta.NewManager(p.DB, p.C.ChartMeta.ReloadInterval, p.C.ChartMeta.Path, p.ChartTrans, p.L)
	err := charts.Init()
	if err != nil {
		return fmt.Errorf("fail to start charts manager: %s", err)
	}

	meta := metricmeta.NewManager(
		p.C.MetricMeta.Sources,
		p.DB,
		p.Index,
		p.C.MetricMeta.MetricMetaPath,
		p.C.MetricMeta.GroupFiles,
		p.MetricTran,
		p.Log,
		p.Redis,
		p.C.MetricMeta.MetricMetaCacheExpiration,
		p.CkMetaLoader,
	)
	err = meta.Init()
	if err != nil {
		return err
	}

	p.q = &Metricq{
		Queryer:         query.New(p.CkMetaLoader, p.Storage, p.CkStorageReader, p.Log),
		queryv1:         queryv1.New(&query.MetricIndexLoader{Interface: p.Index}, charts, p.Meta, p.ChartTrans),
		index:           p.Index,
		meta:            p.Meta,
		charts:          charts,
		handler:         p.queryMetrics,
		handlerV1:       p.queryMetricsV1,
		externalHandler: p.queryExternalMetrics,
	}
	Q = p.q

	routes := ctx.Service("http-server", interceptors.Recover(p.L), interceptors.CORS(true)).(httpserver.Router)
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
