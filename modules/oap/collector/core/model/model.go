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
	lpb "github.com/erda-project/erda-proto-go/oap/logs/pb"
	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	tpb "github.com/erda-project/erda-proto-go/oap/trace/pb"
	"github.com/erda-project/erda/modules/core/monitor/metric"
)

type DataType string

const (
	MetricDataType DataType = "metric"
	TraceDataType  DataType = "trace"
	LogDataType    DataType = "log"
)

type ObservableData interface {
	Clone() ObservableData
	RangeTagsFunc(handle func(tags map[string]string) map[string]string)
	RangeNameFunc(handle func(name string) string)
	SourceData() interface{}
	CompatibilitySourceData() interface{}
}

// Metrics
type Metrics struct {
	Metrics []*mpb.Metric `json:"metrics"`
}

func (m *Metrics) RangeNameFunc(handle func(name string) string) {
	for _, item := range m.Metrics {
		item.Name = handle(item.Name)
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

func (m *Metrics) RangeTagsFunc(handle func(tags map[string]string) map[string]string) {
	for _, item := range m.Metrics {
		item.Attributes = handle(item.Attributes)
	}
}

func (m *Metrics) Clone() ObservableData {
	data := make([]*mpb.Metric, len(m.Metrics))
	copy(data, m.Metrics)
	return &Metrics{Metrics: data}
}

// Traces
type Traces struct {
	Spans []*tpb.Span `json:"spans"`
}

func (t *Traces) RangeNameFunc(handle func(name string) string) {
	for _, item := range t.Spans {
		item.Name = handle(item.Name)
	}
}

func (t *Traces) SourceData() interface{} {
	return t.Spans
}

func (t *Traces) CompatibilitySourceData() interface{} {
	// todo
	return nil
}

func (t *Traces) Clone() ObservableData {
	data := make([]*tpb.Span, len(t.Spans))
	copy(data, t.Spans)
	return &Traces{Spans: data}
}

func (t *Traces) RangeTagsFunc(handle func(tags map[string]string) map[string]string) {
	for _, item := range t.Spans {
		item.Attributes = handle(item.Attributes)
	}
}

// Logs
type Logs struct {
	Logs []*lpb.Log `json:"logs"`
}

func (l *Logs) RangeNameFunc(handle func(name string) string) {
	for _, item := range l.Logs {
		item.Name = handle(item.Name)
	}
}

func (l *Logs) SourceData() interface{} {
	return l.Logs
}

func (l *Logs) CompatibilitySourceData() interface{} {
	// todo
	return nil
}

func (l *Logs) Clone() ObservableData {
	data := make([]*lpb.Log, len(l.Logs))
	copy(data, l.Logs)
	return &Logs{Logs: data}
}

func (l *Logs) RangeTagsFunc(handle func(tags map[string]string) map[string]string) {
	for _, item := range l.Logs {
		item.Attributes = handle(item.Attributes)
	}
}
