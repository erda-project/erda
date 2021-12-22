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
)

type DataType string

const (
	MetricDataType DataType = "metric"
	TraceDataType  DataType = "trace"
	LogDataType    DataType = "log"
)

type ObservableData interface {
	DataType() DataType
	Clone() ObservableData
}

type Metrics struct {
	Metrics []*mpb.Metric `json:"metrics"`
}

func (m *Metrics) Clone() ObservableData {
	data := make([]*mpb.Metric, len(m.Metrics))
	copy(data, m.Metrics)
	return &Metrics{Metrics: data}
}

func (m *Metrics) DataType() DataType {
	return MetricDataType
}

type Traces struct {
	Spans []*tpb.Span `json:"spans"`
}

func (t *Traces) DataType() DataType {
	return TraceDataType
}

func (t *Traces) Clone() ObservableData {
	data := make([]*tpb.Span, len(t.Spans))
	copy(data, t.Spans)
	return &Traces{Spans: data}
}

type Logs struct {
	Logs []*lpb.Log `json:"logs"`
}

func (l *Logs) DataType() DataType {
	return LogDataType
}

func (l *Logs) Clone() ObservableData {
	data := make([]*lpb.Log, len(l.Logs))
	copy(data, l.Logs)
	return &Logs{Logs: data}
}
