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
	"sync"

	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	"github.com/erda-project/erda/modules/core/monitor/metric"
)

// Metrics
type Metrics []*Metric

type Metric struct {
	Meta *Metadata `json:"meta"`
	Data map[string]interface{}
	sync.RWMutex
}

func NewMetric(item *mpb.Metric) *Metric {
	return &Metric{Data: metricToMap(item), Meta: NewMetadata()}
}

func (m *Metric) HandleKeyValuePair(handler func(map[string]interface{}) map[string]interface{}) {
	m.Data = handler(m.Data)
}

func (m *Metric) Name() string {
	return m.Data[NameKey].(string)
}

func (m *Metric) Pairs() map[string]interface{} {
	return m.Data
}

func (m *Metric) Metadata() *Metadata {
	return m.Meta
}

func (m *Metric) Clone() ObservableData {
	m.RLock()
	defer m.RUnlock()
	res := make(map[string]interface{}, len(m.Data))
	for k, v := range m.Data {
		res[k] = v
	}
	return &Metric{
		Data: res,
		Meta: m.Meta.Clone(),
	}
}

func (m *Metric) Source() interface{} {
	return mapToMetric(m.Data)
}

func (m *Metric) SourceCompatibility() interface{} {
	item := mapToMetric(m.Data)
	old := &metric.Metric{}
	old.Timestamp = int64(item.GetTimeUnixNano())
	old.Name = item.GetName()
	old.Tags = item.GetAttributes()
	fields := make(map[string]interface{}, len(item.GetDataPoints()))
	for k, v := range item.DataPoints {
		fields[k] = v
	}
	old.Fields = fields
	return old
}

func (m *Metric) SourceType() SourceType {
	return MetricType
}

func (m *Metric) String() string {
	buf, _ := json.Marshal(m.Data)
	return fmt.Sprintf(string(buf))
}
