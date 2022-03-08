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

package odata

import (
	"encoding/json"
	"fmt"

	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	"github.com/erda-project/erda/modules/core/monitor/metric"
)

// Metrics
type Metrics []*Metric

type Metric struct {
	Item *mpb.Metric `json:"item"`
	Meta *Metadata   `json:"meta"`
}

func NewMetric(item *mpb.Metric) *Metric {
	return &Metric{Item: item, Meta: &Metadata{Data: map[string]string{}}}
}

func (m *Metric) AddMetadata(key, value string) {
	m.Meta.Add(key, value)
}

func (m *Metric) GetMetadata(key string) (string, bool) {
	return m.Meta.Get(key)
}

func (m *Metric) HandleAttributes(handle func(attr map[string]string) map[string]string) {
	m.Item.Attributes = handle(m.Item.Attributes)
}

func (m *Metric) HandleName(handle func(name string) string) {
	m.Item.Name = handle(m.Item.Name)
}

func (m *Metric) Clone() ObservableData {
	item := &mpb.Metric{
		TimeUnixNano: m.Item.TimeUnixNano,
		Name:         m.Item.Name,
		Attributes:   m.Item.Attributes,
		Relations:    m.Item.Relations,
		DataPoints:   m.Item.DataPoints,
	}
	return &Metric{
		Item: item,
		Meta: m.Meta.Clone(),
	}
}

func (m *Metric) Source() interface{} {
	return m.Item
}

func (m *Metric) SourceCompatibility() interface{} {
	old := &metric.Metric{}
	old.Timestamp = int64(m.Item.GetTimeUnixNano())
	old.Name = m.Item.GetName()
	old.Tags = m.Item.GetAttributes()
	fields := make(map[string]interface{}, len(m.Item.GetDataPoints()))
	for k, v := range m.Item.DataPoints {
		fields[k] = v
	}
	old.Fields = fields
	return old
}

func (m *Metric) SourceType() SourceType {
	return MetricType
}

func (m *Metric) String() string {
	buf, _ := json.Marshal(m.Item)
	return fmt.Sprintf("Item => %s", string(buf))
}
