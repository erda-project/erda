package oapspan

import (
	"fmt"
	"strings"
	"sync"

	oap "github.com/erda-project/erda-proto-go/oap/trace/pb"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/common"
	"github.com/erda-project/erda/internal/tools/monitor/oap/collector/lib/common/unmarshalwork"
)

func ParseOapSpanEvent(buf []byte, callback func(event []metric.Metric) error) error {
	uw := newUnmarshalEventWork(buf, callback)
	uw.wg.Add(1)
	unmarshalwork.Schedule(uw)
	uw.wg.Wait()
	if uw.err != nil {
		return fmt.Errorf("parse oapSpan event err: %w", uw.err)
	}
	return nil
}

type unmarshalEventWork struct {
	buf      []byte
	err      error
	wg       sync.WaitGroup
	callback func(event []metric.Metric) error
}

func newUnmarshalEventWork(buf []byte, callback func(event []metric.Metric) error) *unmarshalEventWork {
	return &unmarshalEventWork{buf: buf, callback: callback}
}

func (uw *unmarshalEventWork) Unmarshal() {
	defer uw.wg.Done()
	data := &oap.Span{}
	if err := common.JsonDecoder.Unmarshal(uw.buf, data); err != nil {
		uw.err = fmt.Errorf("json umarshal failed: %w", err)
		return
	}

	if data.Events == nil || len(data.Events) <= 0 {
		return
	}

	eventMetric := metric.Metric{
		Name:    "apm_span_event",
		OrgName: data.Attributes[lib.OrgNameKey],
		Tags:    data.Attributes,
		Fields:  make(map[string]interface{}),
	}
	eventMetric.Tags["_metric_scope_id"] = data.Attributes["terminus_key"]
	eventMetric.Fields["field_count"] = len(data.Events)

	var metrics []metric.Metric

	for _, event := range data.Events {
		eventMetric.Timestamp = int64(event.TimeUnixNano)
		eventMetric.Tags["event"] = event.Name

		if len(event.Attributes) <= 0 {
			continue
		}
		sb := strings.Builder{}
		for k, v := range event.Attributes {
			eventMetric.Tags[k] = v
			sb.WriteString(fmt.Sprintf("%s=%s;", k, v))
		}
		eventMetric.Tags["message"] = sb.String()
		metrics = append(metrics, eventMetric)
	}

	if err := uw.callback(metrics); err != nil {
		uw.err = err
	}
}
