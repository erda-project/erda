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

package promremotewrite

import (
	"fmt"
	"math"
	"net/http"
	"time"

	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	"github.com/erda-project/erda/modules/oap/collector/common"
	"github.com/erda-project/erda/modules/oap/collector/core/model/odata"
	"github.com/golang/protobuf/proto"
	"github.com/labstack/echo"
	pmodel "github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/httpserver"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/erda-project/erda/modules/oap/collector/plugins"
)

var providerName = plugins.WithPrefixReceiver("prometheus-remote-write")

type config struct {
}

// +provider
type provider struct {
	Cfg    *config
	Log    logs.Logger
	Router httpserver.Router `autowired:"http-router"`

	consumerFunc model.ObservableDataConsumerFunc
}

// Run this is optional
func (p *provider) Init(ctx servicehub.Context) error {
	p.Router.POST("/api/v1/collect/prometheus-remote-write", p.prwHandler)
	return nil
}

func (p *provider) prwHandler(ctx echo.Context) error {
	if p.consumerFunc == nil {
		return ctx.NoContent(http.StatusOK)
	}
	req := ctx.Request()
	buf, err := common.ReadBody(req)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("read body err: %s", err))
	}

	var wr prompb.WriteRequest
	err = proto.Unmarshal(buf, &wr)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("unmarshal body err: %s", err))
	}

	now := time.Now()
	for _, ts := range wr.Timeseries {
		attrs := map[string]string{}
		for _, l := range ts.Labels {
			attrs[l.Name] = l.Value
		}
		metricName := attrs[pmodel.MetricNameLabel]
		if metricName == "" {
			return fmt.Errorf("metric name %q not found in attrs or empty", pmodel.MetricNameLabel)
		}
		delete(attrs, pmodel.MetricNameLabel)
		job := attrs[pmodel.JobLabel]
		if metricName == "" {
			return fmt.Errorf("job %q not found in attrs or empty", pmodel.MetricNameLabel)
		}
		delete(attrs, pmodel.JobLabel)
		for _, s := range ts.Samples {
			dataPoints := make(map[string]*structpb.Value)
			if !math.IsNaN(s.Value) {
				dataPoints[metricName] = structpb.NewNumberValue(s.Value)
			}
			// converting to metric
			if len(dataPoints) > 0 {
				t := now
				if s.Timestamp > 0 {
					t = time.Unix(0, s.Timestamp*1000000)
				}
				m := &mpb.Metric{
					Name:         "prw_" + job,
					TimeUnixNano: uint64(t.UnixNano()),
					Attributes:   attrs,
					DataPoints:   dataPoints,
				}
				p.consumerFunc(odata.NewMetric(m))
			}
		}
	}

	return ctx.NoContent(http.StatusOK)
}

func (p *provider) RegisterConsumer(consumer model.ObservableDataConsumerFunc) {
	p.consumerFunc = consumer
}

func (p *provider) ComponentID() model.ComponentID {
	return model.ComponentID(providerName)
}

func init() {
	servicehub.Register(providerName, &servicehub.Spec{
		Services: []string{
			providerName,
		},
		Description: "here is description of prometheus-remote-write",
		ConfigFunc: func() interface{} {
			return &config{}
		},
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
