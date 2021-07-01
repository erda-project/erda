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
	"context"
	"fmt"

	"github.com/recallsong/go-utils/encoding/jsonx"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	indexmanager "github.com/erda-project/erda/modules/core/monitor/metric/index
)

type provider struct {
	Log    logs.Logger
	Index  indexmanager.Index         `autowired:"erda.core.monitor.metric.index"`
	Meta   pb.MetricMetaServiceServer `autowired:"erda.core.monitor.metric.MetricMetaService"`
	Metric pb.MetricServiceServer     `autowired:"erda.core.monitor.metric.MetricService"`
}

func (p *provider) Init(ctx servicehub.Context) error { return nil }

func (p *provider) Run(ctx context.Context) error {
	p.Index.WaitIndicesLoad() // example query can be carried out after the index is loaded
	return p.queryExample(ctx)
}

func (p *provider) queryExample(ctx context.Context) error {
	req := &pb.QueryWithInfluxFormatRequest{
		Start:     "before_1h", // or timestamp
		End:       "now",       // or timestamp
		Statement: `SELECT host_ip::tag, mem_used::field FROM host_summary WHERE cluster_name::tag=$cluster_name GROUP BY host_ip::tag`,
		Params: map[string]*structpb.Value{
			"cluster_name": structpb.NewStringValue("terminus-dev"),
		},
	}
	resp, err := p.Metric.QueryWithInfluxFormat(ctx, req)
	if err != nil {
		return err
	}
	fmt.Println(jsonx.MarshalAndIndent(resp))
	return nil
}

func init() {
	servicehub.Register("metric-query-example", &servicehub.Spec{
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
