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

package spotspan

import (
	"errors"
	"fmt"
	"sync"

	"github.com/erda-project/erda/internal/apps/msp/apm/trace"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/common"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/common/unmarshalwork"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/typeconvert"
)

func ParseSpotSpan(buf []byte, callback func(span *trace.Span) error) error {
	uw := newUnmarshalWork(buf, callback)
	uw.wg.Add(1)
	unmarshalwork.Schedule(uw)
	uw.wg.Wait()
	if uw.err != nil {
		return fmt.Errorf("parse spotSpan err: %w", uw.err)
	}
	return nil
}

// metricToSpan .
func metricToSpan(metric *metric.Metric) (*trace.Span, error) {
	var span trace.Span
	span.Tags = metric.Tags

	traceID, ok := metric.Tags["trace_id"]
	if !ok {
		return nil, errors.New("trace_id cannot be null")
	}
	span.TraceId = traceID
	delete(metric.Tags, "trace_id")

	spanID, ok := metric.Tags["span_id"]
	if !ok {
		return nil, errors.New("span_id cannot be null")
	}
	span.SpanId = spanID
	delete(metric.Tags, "span_id")

	parentSpanID, _ := metric.Tags["parent_span_id"]
	span.ParentSpanId = parentSpanID
	delete(metric.Tags, "parent_span_id")

	value, ok := metric.Fields["start_time"]
	if !ok {
		return nil, errors.New("start_time cannot be null")
	}
	startTime, err := typeconvert.ToInt64(value)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time: %s", value)
	}
	span.StartTime = startTime
	delete(metric.Tags, "start_time")

	value, ok = metric.Fields["end_time"]
	if !ok {
		return nil, errors.New("end_time cannot be null")
	}
	endTime, err := typeconvert.ToInt64(value)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time: %s", value)
	}
	span.EndTime = endTime
	delete(metric.Tags, "end_time")

	return &span, nil
}

type unmarshalWork struct {
	buf      []byte
	err      error
	callback func(span *trace.Span) error
	wg       sync.WaitGroup
}

func newUnmarshalWork(buf []byte, callback func(span *trace.Span) error) *unmarshalWork {
	return &unmarshalWork{buf: buf, callback: callback}
}

func (uw *unmarshalWork) Unmarshal() {
	defer uw.wg.Done()
	data := &metric.Metric{}
	if err := common.JsonDecoder.Unmarshal(uw.buf, data); err != nil {
		uw.err = fmt.Errorf("json umarshal failed: %w", err)
		return
	}
	span, err := metricToSpan(data)
	if err != nil {
		uw.err = fmt.Errorf("cannot convert metric to span: %w", err)
		return
	}
	if v, ok := span.Tags[lib.OrgNameKey]; ok {
		span.OrgName = v
	} else {
		uw.err = fmt.Errorf("must have %q", lib.OrgNameKey)
		return
	}

	if err := uw.callback(span); err != nil {
		uw.err = err
	}
	return
}
