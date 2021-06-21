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

package apis

import (
	"fmt"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/cassandra"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda-infra/providers/httpserver/interceptors"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda-infra/providers/mysql"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/adapt"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/cql"
	"github.com/erda-project/erda/modules/monitor/alert/alert-apis/db"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
	block "github.com/erda-project/erda/modules/monitor/dashboard/chart-block"
	"github.com/erda-project/erda/modules/pkg/bundle-ex/cmdb"
	"github.com/erda-project/erda/pkg/http/httpclient"
)

type define struct{}

func (d *define) Service() []string { return []string{"alert-apis"} }
func (d *define) Dependencies() []string {
	return []string{"http-server", "metrics-query", "mysql", "i18n"}
}
func (d *define) Summary() string     { return "alert apis" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider { return &provider{} }
}

type config struct {
	OrgFilterTags               string `file:"org_filter_tags"`
	MicroServiceFilterTags      string `file:"micro_service_filter_tags"`
	MicroServiceOtherFilterTags string `file:"micro_service_other_filter_tags"`
	SilencePolicy               string `file:"silence_policy"`
	Cassandra                   struct {
		cassandra.SessionConfig `file:"session"`
		GCGraceSeconds          int `file:"gc_grace_seconds" default:"86400"`
	} `file:"cassandra"`
}

type provider struct {
	C                           *config
	L                           logs.Logger
	metricq                     metricq.Queryer
	t                           i18n.Translator
	db                          *db.DB
	cql                         *cql.Cql
	a                           *adapt.Adapt
	bdl                         *bundle.Bundle
	cmdb                        *cmdb.Cmdb
	silencePolicies             map[string]bool
	orgFilterTags               map[string]bool
	microServiceFilterTags      map[string]bool
	microServiceOtherFilterTags map[string]bool
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.silencePolicies = make(map[string]bool)
	p.orgFilterTags = make(map[string]bool)
	p.microServiceFilterTags = make(map[string]bool)
	p.microServiceOtherFilterTags = make(map[string]bool)
	for _, k := range strings.Split(p.C.OrgFilterTags, ",") {
		k = strings.TrimSpace(k)
		if len(k) > 0 {
			p.orgFilterTags[k] = true
		}
	}
	for _, k := range strings.Split(p.C.MicroServiceFilterTags, ",") {
		k = strings.TrimSpace(k)
		if len(k) > 0 {
			p.microServiceFilterTags[k] = true
		}
	}
	for _, k := range strings.Split(p.C.SilencePolicy, ",") {
		k = strings.TrimSpace(k)
		if len(k) > 0 {
			p.silencePolicies[k] = true
		}
	}
	for _, k := range strings.Split(p.C.MicroServiceOtherFilterTags, ",") {
		k = strings.TrimSpace(k)
		if len(k) > 0 {
			p.microServiceOtherFilterTags[k] = true
		}
	}
	cassandra := ctx.Service("cassandra").(cassandra.Interface)
	session, err := cassandra.Session(&p.C.Cassandra.SessionConfig)
	if err != nil {
		return fmt.Errorf("fail to create cassandra session: %s", err)
	}
	p.cql = cql.New(session)
	if err := p.cql.Init(p.L, p.C.Cassandra.GCGraceSeconds); err != nil {
		return fmt.Errorf("fail to init cassandra: %s", err)
	}

	p.t = ctx.Service("i18n").(i18n.I18n).Translator("alert")
	p.db = db.New(ctx.Service("mysql").(mysql.Interface).DB())
	p.metricq = ctx.Service("metrics-query").(metricq.Queryer)
	hc := httpclient.New(httpclient.WithTimeout(time.Second, time.Second*60))
	p.cmdb = cmdb.New(cmdb.WithHTTPClient(hc))
	p.bdl = bundle.New(
		bundle.WithHTTPClient(hc),
		bundle.WithCMDB(),
	)

	dashapi := ctx.Service("chart-block").(block.DashboardAPI)
	p.a = adapt.New(p.L, p.metricq, p.t, p.db, p.cql, p.bdl, p.cmdb, dashapi, p.orgFilterTags, p.microServiceFilterTags, p.microServiceOtherFilterTags, p.silencePolicies)
	routes := ctx.Service("http-server",
		//telemetry.HttpMetric(),
		interceptors.Recover(p.L)).(httpserver.Router)
	return p.intRoutes(routes)
}

func init() {
	servicehub.RegisterProvider("alert-apis", &define{})
}
