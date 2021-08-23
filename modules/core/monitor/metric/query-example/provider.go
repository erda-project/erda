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

package example

import (
	"context"
	"fmt"

	"github.com/recallsong/go-utils/encoding/jsonx"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	indexmanager "github.com/erda-project/erda/modules/core/monitor/metric/index"
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
