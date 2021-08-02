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

package orgapis

import (
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/core/monitor/metric/query/metricq"
	"github.com/erda-project/erda/modules/pkg/bundle-ex/cmdb"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type config struct {
	OfflineTimeout time.Duration `file:"offline_timeout"`
	OfflineSleep   time.Duration `file:"offline_sleep"`
}

type provider struct {
	C       *config
	L       logs.Logger
	bundle  *bundle.Bundle
	cmdb    *cmdb.Cmdb
	metricq metricq.Queryer
	service queryServiceImpl
	t       i18n.Translator
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.t = ctx.Service("i18n").(i18n.I18n).Translator("org-resource")
	hc := httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	p.bundle = bundle.New(
		bundle.WithHTTPClient(hc),
		bundle.WithCoreServices(),
		bundle.WithClusterManager(),
	)
	p.cmdb = cmdb.New(cmdb.WithHTTPClient(hc))
	p.metricq = ctx.Service("metrics-query").(metricq.Queryer)
	p.service = &queryService{metricQ: p.metricq}
	routes := ctx.Service("http-server", interceptors.Recover(p.L)).(httpserver.Router)
	return p.intRoutes(routes)
}

func init() {
	servicehub.Register("org-apis", &servicehub.Spec{
		Services:     []string{"org-apis"},
		Dependencies: []string{"http-server", "metrics-query", "i18n"},
		Description:  "org apis",
		ConfigFunc:   func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
