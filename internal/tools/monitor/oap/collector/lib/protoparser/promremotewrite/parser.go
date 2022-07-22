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

package promremotewrite

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/golang/snappy"
	pmodel "github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
)

func ParseStream(r io.Reader, callback func(record *metric.Metric) error) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	buf, err = snappy.Decode(nil, buf)
	if err != nil {
		return err
	}

	wr := &prompb.WriteRequest{}
	err = wr.Unmarshal(buf)
	if err != nil {
		return fmt.Errorf("unmarshal WriteRequest: %w", err)
	}

	now := time.Now() // receive time
	for _, ts := range wr.Timeseries {
		tags := map[string]string{}
		for _, l := range ts.Labels {
			tags[l.Name] = l.Value
		}
		metricName := tags[pmodel.MetricNameLabel]
		if metricName == "" {
			return fmt.Errorf("%q not found in tags or empty", pmodel.MetricNameLabel)
		}
		delete(tags, pmodel.MetricNameLabel)

		// set pmodel.JobLabel as  name
		job := tags[pmodel.JobLabel]
		if job == "" {
			return fmt.Errorf("%q not found in tags or empty", pmodel.JobLabel)
		}
		delete(tags, pmodel.JobLabel)

		for _, s := range ts.Samples {
			fields := make(map[string]interface{})
			if math.IsNaN(s.Value) {
				continue
			}
			fields[metricName] = s.Value

			// converting to metric
			t := now
			if s.Timestamp > 0 {
				t = time.Unix(0, s.Timestamp*time.Millisecond.Nanoseconds())
			}
			m := metric.Metric{
				Name:      job,
				Timestamp: t.UnixNano(),
				Tags:      tags,
				Fields:    fields,
			}
			err := callback(&m)
			if err != nil {
				return fmt.Errorf("callback: %w", err)
			}
		}
	}
	return nil
}
