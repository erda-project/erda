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

package storage

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	metrics "github.com/erda-project/erda/modules/core/monitor/metric"
)

// TestMetricToSpan .
func TestMetricToSpan(t *testing.T) {
	metric := metrics.Metric{
		Name:      "span",
		Timestamp: time.Now().UnixNano(),
		Tags: map[string]string{"_lt": "transient", "test_tag": "test", "operation_name": "component",
			"parent_span_id": "test-parent-trance-id", "span_id": "test-span-id", "trace_id": "test-trace-id"},
		Fields: map[string]interface{}{"start_time": time.Now().Add(-1).UnixNano(), "end_time": time.Now().UnixNano(),
			"_lt": "transient", "operation_name": "component", "parent_span_id": "test-parent-trance-id",
			"span_id": "test-span-id", "test_tag": "test", "trace_id": "test-trace-id"},
	}
	span, err := metricToSpan(&metric)
	if err != nil {
		log.Fatalln(err)
	}
	if assert.NotNil(t, span, "not nil") {

		assert.Equal(t, metric.Tags["trace_id"], "test-trace-id")
		assert.Equal(t, metric.Tags["parent_span_id"], "test-parent-trance-id")
		assert.Equal(t, metric.Tags["operation_name"], "component")
		assert.Equal(t, metric.Tags["span_id"], "test-span-id")
	}
	fmt.Println(metric)
}

// TestToInt64 .
func TestToInt64(t *testing.T) {
	_int := 10
	_int8 := int8(10)
	_int16 := int16(10)
	_int32 := int32(10)
	_int64 := int64(10)
	_uint := 10
	_uint8 := uint8(10)
	_uint16 := uint16(10)
	_uint32 := uint32(10)
	_uint64 := uint64(10)
	_float32 := float32(10)
	_float64 := float64(10)
	_string := "10"

	num, err := toInt64(_int)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_int8)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_int16)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_int32)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_int64)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_uint)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_uint8)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_uint16)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_uint32)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_uint64)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_float32)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_float64)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))

	num, err = toInt64(_string)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, num, int64(10))
}

// TestToSpan .
func TestToSpan(t *testing.T) {
	span := Span{
		TraceID:       "test-trace-id",
		StartTime:     time.Now().Add(-1).UnixNano(),
		SpanID:        "test-span-id",
		ParentSpanID:  "test-parent-trance-id",
		OperationName: "component",
		EndTime:       time.Now().UnixNano(),
		Tags:          map[string]string{"test_tag": "test"},
	}
	metric := toSpan(&span)
	if assert.NotNil(t, metric, "not nil") {

		assert.Equal(t, metric.Name, "span")
		assert.Equal(t, metric.Timestamp, span.StartTime)
		assert.Equal(t, metric.Tags["_lt"], "transient")
	}
	fmt.Println(metric)
}
