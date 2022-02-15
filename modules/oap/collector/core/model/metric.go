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

package model

import (
	"encoding/json"
	"fmt"

	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	"github.com/erda-project/erda/modules/core/monitor/metric"
)

// Metrics
type Metrics struct {
	Metrics []*mpb.Metric `json:"metrics"`
}

func (m *Metrics) String() string {
	buf, _ := json.Marshal(m.Metrics)
	return fmt.Sprintf("metrics => %s", string(buf))
}

func (m *Metrics) RangeFunc(handle handleFunc) {
	droppedList := make([]int, 0)
	for idx, item := range m.Metrics {
		drop, res := handle(&DataItem{
			TimestampNano: item.TimeUnixNano,
			Name:          item.Name,
			Tags:          item.Attributes,
			Fields:        item.DataPoints,
			Type:          MetricDataType,
		})
		item.Name = res.Name
		item.Attributes = res.Tags
		item.DataPoints = res.Fields
		if drop {
			// TODO mark dropped item
			droppedList = append(droppedList, idx)
		}
	}
}

func (m *Metrics) RangeNameFunc(handle func(name string) string) {
	for _, item := range m.Metrics {
		item.Name = handle(item.Name)
	}
}

func (m *Metrics) RangeTagsFunc(handle func(tags map[string]string) map[string]string) {
	for _, item := range m.Metrics {
		item.Attributes = handle(item.Attributes)
	}
}

func (m *Metrics) SourceData() interface{} {
	return m.Metrics
}

func (m *Metrics) CompatibilitySourceData() interface{} {
	res := make([]*metric.Metric, len(m.Metrics))
	for idx, item := range m.Metrics {
		fields := make(map[string]interface{}, len(item.DataPoints))
		for k, v := range item.DataPoints {
			fields[k] = v
		}
		res[idx] = &metric.Metric{
			Name:      item.Name,
			Timestamp: int64(item.TimeUnixNano),
			Tags:      item.Attributes,
			Fields:    fields,
		}
	}
	return map[string]interface{}{
		"metrics": res,
	}
}

func (m *Metrics) Clone() ObservableData {
	data := make([]*mpb.Metric, len(m.Metrics))
	copy(data, m.Metrics)
	return &Metrics{Metrics: data}
}
