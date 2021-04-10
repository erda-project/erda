// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package storage

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/erda-project/erda/modules/monitor/core/metrics"
	"github.com/erda-project/erda/modules/monitor/trace"
	"github.com/stretchr/testify/assert"
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
	span := trace.Span{
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
