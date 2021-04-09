// Copyright (c) 2021 Terminus, Inc.

// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later (AGPL), as published by the Free Software Foundation.

// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.

// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package example

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/recallsong/go-utils/encoding/jsonx"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	indexmanager "github.com/erda-project/erda/modules/monitor/core/metrics/index"
	"github.com/erda-project/erda/modules/monitor/core/metrics/metricq"
)

type define struct{}

func (d *define) Service() []string { return []string{"metricq-example"} }
func (d *define) Dependencies() []string {
	return []string{"telemetry-query"}
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
	C       *config
	L       logs.Logger
	index   indexmanager.Index
	metricq metricq.Queryer
}

func (p *provider) Init(ctx servicehub.Context) error {
	p.index = ctx.Service("telemetry-index-manager").(indexmanager.Index) // 不是必须的，引用只是为了，例子查询能在索引加载后进行
	p.metricq = ctx.Service("telemetry-query").(metricq.Queryer)
	return nil
}

func (p *provider) Start() error {
	p.index.WaitIndicesLoad() // 为了下面的查询能在索引加载后进行
	// p.queryExample()
	// nodesOld, err := p.getNodeInfoV1()
	// if err != nil {
	// 	return err
	// }
	// for _, n := range nodesOld {
	// 	println(n.HostIP)
	// }
	//
	// nodes, err := p.getNodeInfo()
	// if err != nil {
	// 	return err
	// }
	// for _, n := range nodes {
	// 	println(n.HostIP)
	// }

	// nodes := []*cNode{{"10.0.6.505"}}
	// r1, err := p.getNodeDiskUsageV1(nodes[0])
	// r2, err2 := p.getNodeDiskUsage(nodes[0])
	// if err != nil {
	// 	println("r1 error")
	// }
	// if err2 != nil {
	// 	println("r2 error")
	// }
	// println(r1, "-----", r2)

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

func (p *provider) getNodeInfoV1() (res []*cNode, err error) {
	source, err := p.metricq.QueryWithFormatV1("params", "docker_container_status?start=before_30m&end=now&group=tags.host_ip&filter_tags.addon_type=cassandra", "raw", "en")
	if err != nil {
		return nil, err
	}
	var data []struct {
		Tag string `mapstructure:"tag"`
	}

	println(reflect.TypeOf(source.Data).String())
	if err := mapstructure.Decode(source.Data, &data); err != nil {
		return nil, err
	}

	for _, d := range data {
		res = append(res, &cNode{HostIP: d.Tag})
	}

	return
}

func (p *provider) getNodeInfo() (res []*cNode, err error) {
	options := url.Values{}
	options.Set("start", "before_30m")
	options.Set("end", "now")
	source, err := p.metricq.Query(metricq.InfluxQL,
		`SELECT host_ip::tag FROM docker_container_status WHERE addon_type::tag='cassandra' GROUP BY host_ip::tag`,
		nil,
		options)
	if err != nil {
		return nil, err
	}
	for _, d := range source.ResultSet.Rows {
		res = append(res, &cNode{HostIP: fmt.Sprintf("%v", d[0])})
	}
	return
}

func (p *provider) getNodeDiskUsageV1(node *cNode) (res float64, err error) {
	source1, err := p.metricq.QueryWithFormatV1("params", fmt.Sprintf("disk?start=before_1h&end=now&filter_tags.host_ip=%s&sort=timestamp&limit=1", node.HostIP), "raw", "en")
	if err != nil {
		return 0, err
	}
	var data []struct {
		Fields struct {
			UsedPercent float64 `mapstructure:"used_percent"`
		} `mapstructure:"fields"`
	}

	if err := mapstructure.Decode(source1.Data, &data); err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("no data. data=%+v", data)
	}
	res = data[0].Fields.UsedPercent
	return
}

func (p *provider) getNodeDiskUsage(node *cNode) (res float64, err error) {
	options := url.Values{}
	options.Set("start", "before_30m")
	options.Set("end", "now")
	//options.Set("debug", "true")
	source, err := p.metricq.Query(metricq.InfluxQL,
		"SELECT used_percent::field FROM disk WHERE host_ip::tag=$host_ip ORDER BY timestamp LIMIT 1",
		map[string]interface{}{
			"host_ip": node.HostIP,
		},
		options)
	if err != nil {
		return 0, err
	}
	if source.Details != nil {
		fmt.Println(source.Details) // debug=true时，打印调试信息
	}
	if source.ResultSet.Rows == nil {
		return 0, fmt.Errorf("no data")
	}
	if len(source.ResultSet.Rows) <= 0 || len(source.ResultSet.Rows[0]) <= 0 {
		return 0, fmt.Errorf("no data")
	}
	res, _ = source.ResultSet.Rows[0][0].(float64)
	return
}

func (p *provider) Close() error { return nil }

func init() {
	servicehub.RegisterProvider("metricq-example", &define{})
}
