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

package table

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/monitor/alert/pb"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/pkg/common/apis"
)

type TriggerCount struct {
	tk           string
	metric       metricpb.MetricServiceServer `autowired:"erda.core.monitor.metric.MetricService"`
	triggerCount map[string]float64
	events       []*pb.AlertEventItem
	ctx          context.Context
}

// fetch step
var step = 20

func New(ctx context.Context, metric metricpb.MetricServiceServer, tk string, events []*pb.AlertEventItem) *TriggerCount {
	metricQueryCtx := apis.GetContext(ctx, func(header *transport.Header) {
	})

	return &TriggerCount{
		tk:           tk,
		metric:       metric,
		events:       events,
		triggerCount: make(map[string]float64),
		ctx:          metricQueryCtx,
	}
}

func (f *TriggerCount) Get(ctx context.Context, item *pb.AlertEventItem) float64 {
	return f.triggerCount[item.Id]
}

func (f *TriggerCount) Fetch() error {
	if len(f.events) <= 0 {
		return nil
	}

	mod := (len(f.events) / step) + 1
	for i := 0; i < mod; i++ {
		l := i * step
		var events []*pb.AlertEventItem
		if l+step > len(f.events) {
			events = f.events[l:]
		} else {
			events = f.events[l : l+step]
		}

		if len(events) <= 0 {
			break
		}

		var ids []interface{}

		for _, event := range events {
			ids = append(ids, event.Id)
		}
		triggerCount, err := fetch(f.ctx, f.metric, ids)
		if err != nil {
			return err
		}

		for id, count := range triggerCount {
			f.triggerCount[id] = count
		}
	}
	return nil
}

func fetch(ctx context.Context, metricService metricpb.MetricServiceServer, ids []interface{}) (map[string]float64, error) {
	params, err := structpb.NewList(ids)
	if err != nil {
		return nil, err
	}

	statement := fmt.Sprintf("SELECT count(timestamp),family_id::tag FROM analyzer_alert WHERE alert_suppressed::tag='false' group by family_id::tag")
	resp, err := metricService.QueryWithInfluxFormat(ctx, &metricpb.QueryWithInfluxFormatRequest{
		Start:     "0",
		End:       strconv.FormatInt(time.Now().UnixNano()/1e6, 10),
		Statement: statement,
		Filters: []*metricpb.Filter{
			{
				Key:   "tags.family_id",
				Op:    "in",
				Value: structpb.NewListValue(params),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Results == nil {
		return nil, nil
	}

	result := make(map[string]float64)

	for _, triggerCount := range resp.Results[0].Series[0].Rows {
		v := triggerCount.GetValues()
		result[v[1].GetStringValue()] = v[0].GetNumberValue()
	}

	return result, nil
}
