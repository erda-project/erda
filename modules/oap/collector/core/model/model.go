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

type ComponentID string

type ObserveData interface {
}

// data trunk of observability
type Metrics struct {
	Metrics []*mpb.Metric
}

func (ms Metrics) Clone() Metrics {
	data := make([]*mpb.Metric, len(ms.Metrics))
	copy(data, ms.Metrics)
	return Metrics{Metrics: data}
}

type Traces struct {
	Spans []*tpb.Span
}

type Logs struct {
	Logs []*lpb.Log
}
