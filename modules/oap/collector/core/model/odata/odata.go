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
	"sort"
	"strings"

	"github.com/cespare/xxhash"
	lpb "github.com/erda-project/erda-proto-go/oap/logs/pb"
	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	tpb "github.com/erda-project/erda-proto-go/oap/trace/pb"
	"github.com/erda-project/erda/modules/oap/collector/common/pbconvert"
	structpb "github.com/golang/protobuf/ptypes/struct"

	"github.com/erda-project/erda-proto-go/oap/common/pb"
)

type SourceType string

const (
	MetricType SourceType = "METRIC"
	SpanType   SourceType = "SPAN"
	LogType    SourceType = "LOG"
	RawType    SourceType = "RAW"
)

// Global internal keywords
const (
	KeyWordPrefix   = "__kw__"
	NameKey         = KeyWordPrefix + "name"
	TimeUnixNanoKey = KeyWordPrefix + "ts" // Type: uint64
	// The keys in Attributes<map[string]string> without any prefix
	// Log
	SeverityKey = KeyWordPrefix + "severity"
	ContentKey  = KeyWordPrefix + "content"
	// Metric
	// Refer to `mpb.Metric.DataPoints` representation
	DataPointsKeyPrefix = "__dp__"
	// Spans
	TraceID           = KeyWordPrefix + "trace_id"
	SpanID            = KeyWordPrefix + "span_id"
	ParentSpanID      = KeyWordPrefix + "parent_span_id"
	StartTimeUnixNano = KeyWordPrefix + "start_ts"
	EndTimeUnixNano   = KeyWordPrefix + "end_ts"
)

type ObservableData interface {
	HandleKeyValuePair(handler func(pairs map[string]interface{}) map[string]interface{})
	Pairs() map[string]interface{}
	Name() string
	Metadata() *Metadata
	Clone() ObservableData
	Source() interface{}
	SourceCompatibility() interface{}
	SourceType() SourceType
	String() string
}

type SourceItem interface {
	GetName() string
	GetAttributes() map[string]string
	GetRelations() *pb.Relation
}

// fieldKey is field of measurement
// certain metric if fields if sourceItem is Metric
// empty string if sourceItem is Log
// ...
func HashSourceItem(fieldKey string, sourceItem SourceItem) uint64 {
	var sb strings.Builder
	keys := make([]string, 0, len(sourceItem.GetAttributes()))
	for k := range sourceItem.GetAttributes() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sb.WriteString(sourceItem.GetName() + "\n")
	for _, k := range keys {
		sb.WriteString(k + sourceItem.GetAttributes()[k] + "\n")
	}
	sb.WriteString(fieldKey)

	return xxhash.Sum64String(sb.String())
}

func extractAttributes(data map[string]interface{}) map[string]string {
	attr := make(map[string]string)
	for k, v := range data {
		if strings.HasPrefix(k, KeyWordPrefix) {
			continue
		}
		sv, ok := v.(string)
		if !ok {
			continue
		}
		attr[k] = sv
	}
	return attr
}

func logToMap(item *lpb.Log) map[string]interface{} {
	data := make(map[string]interface{}, len(item.GetAttributes()))
	data[NameKey] = item.GetName()
	data[TimeUnixNanoKey] = item.GetTimeUnixNano()
	for k, v := range item.GetAttributes() {
		data[k] = v
	}
	data[SeverityKey] = item.GetSeverity()
	data[ContentKey] = item.GetContent()
	return data
}

func mapToLog(data map[string]interface{}) *lpb.Log {
	return &lpb.Log{
		TimeUnixNano: data[TimeUnixNanoKey].(uint64),
		Name:         data[NameKey].(string),
		Severity:     data[SeverityKey].(string),
		Attributes:   extractAttributes(data),
		Content:      data[ContentKey].(string),
	}
}

func metricToMap(item *mpb.Metric) map[string]interface{} {
	data := make(map[string]interface{}, len(item.GetAttributes())+len(item.GetDataPoints()))
	data[NameKey] = item.GetName()
	data[TimeUnixNanoKey] = item.GetTimeUnixNano()
	for k, v := range item.GetAttributes() {
		data[k] = v
	}
	for k, v := range item.GetDataPoints() {
		data[DataPointsKeyPrefix+k] = v.AsInterface()
	}
	return data
}

func mapToMetric(data map[string]interface{}) *mpb.Metric {
	dps := make(map[string]*structpb.Value)
	for k, v := range data {
		if !strings.HasPrefix(k, DataPointsKeyPrefix) {
			continue
		}
		dps[strings.TrimPrefix(k, DataPointsKeyPrefix)] = pbconvert.ToValue(v)
	}
	return &mpb.Metric{
		TimeUnixNano: data[TimeUnixNanoKey].(uint64),
		Name:         data[NameKey].(string),
		Attributes:   extractAttributes(data),
		DataPoints:   dps,
	}
}

func spanToMap(item *tpb.Span) map[string]interface{} {
	data := make(map[string]interface{}, len(item.Attributes))
	data[NameKey] = item.Name
	for k, v := range item.Attributes {
		data[k] = v
	}

	data[TraceID] = item.TraceID
	data[SpanID] = item.SpanID
	data[ParentSpanID] = item.ParentSpanID
	data[StartTimeUnixNano] = item.StartTimeUnixNano
	data[EndTimeUnixNano] = item.EndTimeUnixNano
	return data
}

func mapToSpan(data map[string]interface{}) *tpb.Span {
	return &tpb.Span{
		Name:              data[NameKey].(string),
		Attributes:        extractAttributes(data),
		TraceID:           data[TraceID].(string),
		SpanID:            data[SpanID].(string),
		ParentSpanID:      data[ParentSpanID].(string),
		StartTimeUnixNano: data[StartTimeUnixNano].(uint64),
		EndTimeUnixNano:   data[EndTimeUnixNano].(uint64),
	}
}
