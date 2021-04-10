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

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	indexmanager "github.com/erda-project/erda/modules/monitor/core/metrics/index"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
	"github.com/recallsong/go-utils/encoding/jsonx"
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
type cNode struct {
	HostIP string
}

type provider struct {
	Cfg     *config
	Log     logs.Logger
	index   indexmanager.Index
	metricq metricq.Queryer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.index = ctx.Service("metrics-index-manager").(indexmanager.Index) // 不是必须的，引用只是为了，例子查询能在索引加载后进行
	p.metricq = ctx.Service("metrics-query").(metricq.Queryer)
	return nil
}

func (p *provider) Start() error {
	p.index.WaitIndicesLoad() // 为了下面的查询能在索引加载后进行
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
	// 	fmt.Println(rs.Details) // debug=true时，打印调试信息
	// }
	fmt.Println(jsonx.MarshalAndIntend(rs)) // 打印查询结果
}

func (p *provider) Close() error { return nil }

func init() {
	servicehub.RegisterProvider("metricq-example", &define{})
}
