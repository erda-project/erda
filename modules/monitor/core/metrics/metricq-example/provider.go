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

package example

import (
	"fmt"
	"net/url"

	"github.com/recallsong/go-utils/encoding/jsonx"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	indexmanager "github.com/erda-project/erda/modules/monitor/core/metrics/index"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
)

type define struct{}

func (d *define) Services() []string { return []string{"metricq-example"} }
func (d *define) Dependencies() []string {
	return []string{"metrics-query"}
}
func (d *define) Summary() string     { return "metricq-example" }
func (d *define) Description() string { return d.Summary() }
func (d *define) Config() interface{} { return &config{} }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type config struct {
	// some config field
}

type provider struct {
	Cfg     *config
	Log     logs.Logger
	index   indexmanager.Index
	metricq metricq.Queryer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.index = ctx.Service("metrics-index-manager").(indexmanager.Index) // Not necessary，the reference is only so that the example query can be carried out after the index is loaded.
	p.metricq = ctx.Service("metrics-query").(metricq.Queryer)
	return nil
}

func (p *provider) Start() error {
	p.index.WaitIndicesLoad() // Example query can be carried out after the index is loaded
	p.queryExample()
	return nil
}

func (p *provider) queryExample() {
	options := url.Values{}
	options.Set("start", "before_1h") // or timestamp
	options.Set("end", "now")         // or timestamp
	// options.Set("debug", "true")
	rs, err := p.metricq.Query(
		metricq.InfluxQL,
		`SELECT host_ip::tag, mem_used::field FROM host_summary WHERE cluster_name::tag=$cluster_name GROUP BY host_ip::tag`,
		map[string]interface{}{
			"cluster_name": "terminus-dev",
		},
		options,
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	// if rs.Details != nil {
	// 	fmt.Println(rs.Details) // When debug=true，print debug session .
	// }
	fmt.Println(jsonx.MarshalAndIntend(rs)) // Print debug session .
}

func (p *provider) Close() error { return nil }

func init() {
	servicehub.RegisterProvider("metricq-example", &define{})
}
