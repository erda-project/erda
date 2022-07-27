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

package orgapis

import (
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/clickhouse"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/i18n"

	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/internal/core/org"
	"github.com/erda-project/erda/internal/pkg/bundle-ex/cmdb"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric/query/metricq"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric/storage/elasticsearch"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/loader"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type config struct {
	OfflineTimeout     time.Duration `file:"offline_timeout"`
	OfflineSleep       time.Duration `file:"offline_sleep"`
	QueryMetricsFromCk bool          `file:"query_metric_from_clickhouse"`
	DebugSQL           bool          `file:"debug_sql"`
}

type provider struct {
	C       *config
	L       logs.Logger
	bundle  *bundle.Bundle
	cmdb    *cmdb.Cmdb
	metricq metricq.Queryer
	service queryServiceImpl
	t       i18n.Translator

	ClusterSvc  clusterpb.ClusterServiceServer `autowired:"erda.core.clustermanager.cluster.ClusterService"`
	Clickhouse  clickhouse.Interface           `autowired:"clickhouse" optional:"true"`
	Loader      loader.Interface               `autowired:"clickhouse.table.loader@metric" optional:"true"`
	EsSearchRaw elasticsearch.Interface        `autowired:"metric-storage"`
	Org         org.ClientInterface
	Source      MetricSource
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.t = ctx.Service("i18n").(i18n.I18n).Translator("org-resource")
	hc := httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	p.bundle = bundle.New(
		bundle.WithHTTPClient(hc),
		bundle.WithErdaServer(),
		bundle.WithClusterManager(),
	)
	p.cmdb = cmdb.New(cmdb.WithHTTPClient(hc))
	routes := ctx.Service("http-server", interceptors.Recover(p.L)).(httpserver.Router)
	//if !p.C.QueryMetricsFromCk {
	//	p.metricq = ctx.Service("metrics-query").(metricq.Queryer)
	//	p.service = &queryService{metricQ: p.metricq}
	//	return p.intRoutes(routes)
	//} else {
	//	p.Source = &ClickhouseSource{
	//		p:          p,
	//		Clickhouse: p.Clickhouse,
	//		Log:        p.L,
	//		DebugSQL:   p.C.DebugSQL,
	//		Loader:     p.Loader,
	//	}
	//	return p.initRoutesV2(routes)
	//}
	p.metricq = ctx.Service("metrics-query").(metricq.Queryer)
	p.service = &queryService{metricQ: p.metricq}
	p.Source = &ClickhouseSource{
		p:          p,
		Clickhouse: p.Clickhouse,
		Log:        p.L,
		DebugSQL:   p.C.DebugSQL,
		Loader:     p.Loader,
	}
	return p.intRoutes(routes)
}

func init() {
	servicehub.Register("org-apis", &servicehub.Spec{
		Services:             []string{"org-apis"},
		Dependencies:         []string{"http-server", "i18n"},
		OptionalDependencies: []string{"metrics-query"},
		Description:          "org apis",
		ConfigFunc:           func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
