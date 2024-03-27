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

package opentelemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	pb "github.com/erda-project/erda-proto-go/opentelemetry/proto/common/v1/pb"
	pb1 "github.com/erda-project/erda-proto-go/opentelemetry/proto/metrics/v1/pb"
	pb2 "github.com/erda-project/erda-proto-go/opentelemetry/proto/resource/v1/pb"
	emetric "github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/core/model/odata"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/protoparser/jsonmarshal"
)

type MetricOtlpToErda struct {
	logger logs.Logger
	p      *provider
}

func (c *MetricOtlpToErda) writeMetric(ctx context.Context, resource *pb2.Resource, instrumentationLibrary *pb.InstrumentationScope, metrics []*pb1.Metric) error {
	HeapMemoryMetric := c.initErdaMetric("jvm_memory", "heap_memory", resource, instrumentationLibrary)
	NoHeapMemoryMetric := c.initErdaMetric("jvm_memory", "non_heap_memory", resource, instrumentationLibrary)
	EdenSpaceMemoryMetric := c.initErdaMetric("jvm_memory", "java_eden_space", resource, instrumentationLibrary)
	SurvivorSpaceMemoryMetric := c.initErdaMetric("jvm_memory", "java_survivor_space", resource, instrumentationLibrary)
	OldGenMemoryMetric := c.initErdaMetric("jvm_memory", "java_old_gen", resource, instrumentationLibrary)

	GCCmsMetric := c.initErdaMetric("jvm_gc", "concurrentmarksweep", resource, instrumentationLibrary)
	GCPnMetric := c.initErdaMetric("jvm_gc", "parnew", resource, instrumentationLibrary)

	ClassMetric := c.initErdaMetric("jvm_class_loader", "", resource, instrumentationLibrary)
	ThreadDaemonMetrc := c.initErdaMetric("jvm_thread", "daemon_count", resource, instrumentationLibrary)
	ThreadNoDaemonMetrc := c.initErdaMetric("jvm_thread", "nodaemon_count", resource, instrumentationLibrary)

	for k := range metrics {
		if metrics[k].Name == "process.runtime.jvm.memory.usage" {
			memoryResut := calculateMemoryMetric(metrics[k])
			HeapMemoryMetric.Fields["used"] = memoryResut["heap"]
			NoHeapMemoryMetric.Fields["used"] = memoryResut["no_heap"]
			EdenSpaceMemoryMetric.Fields["used"] = memoryResut["edge_space"]
			SurvivorSpaceMemoryMetric.Fields["used"] = memoryResut["survivor_space"]
			OldGenMemoryMetric.Fields["used"] = memoryResut["old_gen"]
		}
		if metrics[k].Name == "process.runtime.jvm.memory.init" {
			memoryResut := calculateMemoryMetric(metrics[k])
			HeapMemoryMetric.Fields["init"] = memoryResut["heap"]
			NoHeapMemoryMetric.Fields["init"] = memoryResut["no_heap"]
			EdenSpaceMemoryMetric.Fields["init"] = memoryResut["edge_space"]
			SurvivorSpaceMemoryMetric.Fields["init"] = memoryResut["survivor_space"]
			OldGenMemoryMetric.Fields["init"] = memoryResut["old_gen"]

		}
		if metrics[k].Name == "process.runtime.jvm.memory.committed" {
			memoryResut := calculateMemoryMetric(metrics[k])
			HeapMemoryMetric.Fields["committed"] = memoryResut["heap"]
			NoHeapMemoryMetric.Fields["committed"] = memoryResut["no_heap"]
			EdenSpaceMemoryMetric.Fields["committed"] = memoryResut["edge_space"]
			SurvivorSpaceMemoryMetric.Fields["committed"] = memoryResut["survivor_space"]
			OldGenMemoryMetric.Fields["committed"] = memoryResut["old_gen"]

		}
		if metrics[k].Name == "process.runtime.jvm.memory.limit" {
			memoryResut := calculateMemoryMetric(metrics[k])
			HeapMemoryMetric.Fields["max"] = memoryResut["heap"]
			NoHeapMemoryMetric.Fields["max"] = memoryResut["no_heap"]
			EdenSpaceMemoryMetric.Fields["max"] = memoryResut["edge_space"]
			SurvivorSpaceMemoryMetric.Fields["max"] = memoryResut["survivor_space"]
			OldGenMemoryMetric.Fields["max"] = memoryResut["old_gen"]

		}
		if metrics[k].Name == "process.runtime.jvm.gc.duration" {
			gcResult := calculateGcMetric(metrics[k])
			GCCmsMetric.Fields["count"] = gcResult["cms_count"]
			GCCmsMetric.Fields["time"] = gcResult["cms_duration"]
			GCPnMetric.Fields["count"] = gcResult["pn_count"]
			GCPnMetric.Fields["time"] = gcResult["pn_duration"]
		}
		if metrics[k].Name == "process.runtime.jvm.classes.loaded" {
			ClassMetric.Fields["loaded"] = metrics[k].GetSum().DataPoints[0].GetAsInt()
		}
		if metrics[k].Name == "process.runtime.jvm.classes.unloaded" {
			ClassMetric.Fields["unloaded"] = metrics[k].GetSum().DataPoints[0].GetAsInt()
		}
		if metrics[k].Name == "process.runtime.jvm.threads.count" {
			threadResult := calculateThreadMetric(metrics[k])
			ThreadDaemonMetrc.Fields["state"] = threadResult["daemon_count"]
			ThreadNoDaemonMetrc.Fields["state"] = threadResult["nodaemon_count"]

		}

	}
	_ = c.Consumer(HeapMemoryMetric)
	_ = c.Consumer(NoHeapMemoryMetric)
	_ = c.Consumer(EdenSpaceMemoryMetric)
	_ = c.Consumer(SurvivorSpaceMemoryMetric)
	_ = c.Consumer(OldGenMemoryMetric)
	_ = c.Consumer(GCCmsMetric)
	_ = c.Consumer(GCPnMetric)
	_ = c.Consumer(ClassMetric)
	_ = c.Consumer(ThreadDaemonMetrc)
	_ = c.Consumer(ThreadNoDaemonMetrc)

	return nil
}

func calculateThreadMetric(metric *pb1.Metric) map[string]int64 {
	result := make(map[string]int64)

	sum := metric.GetSum()

	for i := 0; i < len(sum.GetDataPoints()); i++ {
		dataPoint := sum.GetDataPoints()[i]
		attr := convertAttributes(dataPoint.Attributes)
		if attr["daemon"] == "true" {
			result["daemon_count"] = int64(math.Abs(float64(dataPoint.GetAsInt())))
		}
		if attr["daemon"] == "false" {
			result["nodaemon_count"] = int64(math.Abs(float64(dataPoint.GetAsInt())))
		}
	}
	return result
}

func calculateGcMetric(metric *pb1.Metric) map[string]int64 {
	result := make(map[string]int64)

	histogram := metric.GetHistogram()

	for i := 0; i < len(histogram.GetDataPoints()); i++ {
		dataPoint := histogram.GetDataPoints()[i]
		attr := convertAttributes(dataPoint.Attributes)
		if attr["gc"] == "ConcurrentMarkSweep" {
			result["cms_count"] = int64(dataPoint.Count)
			result["cms_duration"] = int64(*dataPoint.Sum) / int64(dataPoint.Count)
		}
		if attr["gc"] == "ParNew" {
			result["pn_count"] = int64(dataPoint.Count)
			result["pn_duration"] = int64(*dataPoint.Sum) / int64(dataPoint.Count)
		}
	}
	return result
}

func calculateMemoryMetric(metric *pb1.Metric) map[string]int64 {
	result := make(map[string]int64)
	heap := make(map[string]int64)
	no_heap := make(map[string]int64)

	sum := metric.GetSum()

	for i := 0; i < len(sum.GetDataPoints()); i++ {
		dataPoint := sum.GetDataPoints()[i]
		attr := convertAttributes(dataPoint.Attributes)
		if attr["type"] == "non_heap" {
			no_heap[attr["pool"]] = dataPoint.GetAsInt()
			result["no_heap"] += dataPoint.GetAsInt()
		}
		if attr["type"] == "heap" {
			heap[attr["pool"]] = dataPoint.GetAsInt()
			result["heap"] += dataPoint.GetAsInt()

		}
	}
	result["survivor_space"] = heap["Par Survivor Space"]
	result["edge_space"] = heap["Par Eden Space"]
	result["old_gen"] = heap["CMS Old Gen"]
	return result
}

func convertAttributes(attrs []*pb.KeyValue) map[string]string {
	result := make(map[string]string)
	for i := 0; i < len(attrs); i++ {
		attr := attrs[i]
		vv, _ := AttributeValueToInfluxTagValue(attr.Value)
		result[attr.Key] = vv

	}
	return result
}
func (c *MetricOtlpToErda) initErdaMetric(metric_group string, name string, resource *pb2.Resource, instrumentationLibrary *pb.InstrumentationScope) *emetric.Metric {

	m := new(emetric.Metric)

	tags := make(map[string]string)
	var attrs []*pb.KeyValue

	attrs = append(attrs, instrumentationLibrary.Attributes...)
	attrs = append(attrs, resource.Attributes...)

	for _, i := range attrs {
		if i.Key == "" {
			continue
		}
		var vv string
		vv, _ = AttributeValueToInfluxTagValue(i.Value)
		tags[strings.Replace(i.Key, ".", "_", -1)] = vv
	}
	tags["name"] = name
	tags["_metric_scope_id"] = tags["terminus_key"]
	m.Tags = tags
	m.Fields = make(map[string]interface{})
	m.Name = metric_group
	m.Timestamp = time.Now().UnixNano()
	m.OrgName = tags["org_name"]
	return m
}

func (c MetricOtlpToErda) Consumer(metric *emetric.Metric) error {
	err := jsonmarshal.ParseInterface(metric, func(buf []byte) error {
		od := odata.NewRaw(buf)
		od.Meta[c.p.Cfg.MetadataKeyOfTopic] = "spot-metrics"
		return c.p.consumer(od)
	})
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}
	return nil
}

func AttributeValueToInfluxTagValue(value *pb.AnyValue) (string, error) {
	switch value.Value.(type) {
	case *pb.AnyValue_StringValue:
		return value.GetStringValue(), nil
	case *pb.AnyValue_IntValue:
		return strconv.FormatInt(value.GetIntValue(), 10), nil
	case *pb.AnyValue_DoubleValue:
		return strconv.FormatFloat(value.GetDoubleValue(), 'f', -1, 64), nil
	case *pb.AnyValue_BoolValue:
		return strconv.FormatBool(value.GetBoolValue()), nil
	case *pb.AnyValue_KvlistValue:
		if jsonBytes, err := json.Marshal(otlpKeyValueListToMap(value.GetKvlistValue())); err != nil {
			return "", err
		} else {
			return string(jsonBytes), nil
		}
	case *pb.AnyValue_ArrayValue:
		if jsonBytes, err := json.Marshal(otlpArrayToSlice(value.GetArrayValue())); err != nil {
			return "", err
		} else {
			return string(jsonBytes), nil
		}
	default:
		return "", fmt.Errorf("unknown value type")
	}
}

func otlpKeyValueListToMap(kvList *pb.KeyValueList) map[string]interface{} {
	m := make(map[string]interface{}, len(kvList.Values))
	for _, v := range kvList.Values {
		switch v.Value.Value.(type) {
		case *pb.AnyValue_StringValue:
			m[v.Key] = v.Value.GetStringValue()
		case *pb.AnyValue_IntValue:
			m[v.Key] = v.Value.GetIntValue()
		case *pb.AnyValue_DoubleValue:
			m[v.Key] = v.Value.GetDoubleValue()
		case *pb.AnyValue_BoolValue:
			m[v.Key] = v.Value.GetBoolValue()
		case *pb.AnyValue_KvlistValue:
			m[v.Key] = otlpKeyValueListToMap(v.Value.GetKvlistValue())
		case *pb.AnyValue_ArrayValue:
			m[v.Key] = otlpArrayToSlice(v.Value.GetArrayValue())
		default:
			m[v.Key] = fmt.Sprintf("<invalid map value> %v", v)
		}
	}
	return m
}

func otlpArrayToSlice(arr *pb.ArrayValue) []interface{} {
	s := make([]interface{}, 0, len(arr.Values))
	for _, v := range arr.Values {
		switch v.Value.(type) {
		case *pb.AnyValue_StringValue:
			s = append(s, v.GetStringValue())
		case *pb.AnyValue_IntValue:
			s = append(s, v.GetIntValue())
		case *pb.AnyValue_DoubleValue:
			s = append(s, v.GetDoubleValue())
		case *pb.AnyValue_BoolValue:
			s = append(s, v.GetBoolValue())
		default:
			s = append(s, fmt.Sprintf("<invalid array value> %v", v))
		}
	}
	return s
}
