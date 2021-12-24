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
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	mpb "github.com/erda-project/erda-proto-go/oap/metrics/pb"
	"github.com/erda-project/erda/modules/oap/collector/core/model"
	"github.com/golang/snappy"
	pmodel "github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"google.golang.org/protobuf/types/known/structpb"
)

func ReadBody(req *http.Request) ([]byte, error) {
	encoding := req.Header.Get("Content-Encoding")
	defer req.Body.Close()

	switch encoding {
	case "gzip":
		r, err := gzip.NewReader(req.Body)
		if err != nil {
			return nil, fmt.Errorf("gzip.NewReader err: %w", err)
		}

		bytes, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	case "snappy":
		bytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		// snappy block format is only supported by decode/encode not snappy reader/writer
		bytes, err = snappy.Decode(nil, bytes)
		if err != nil {
			return nil, fmt.Errorf("snappy.Decode err: %w", err)
		}
		return bytes, nil
	default:
		bytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	}
}

func convertToMetrics(wr prompb.WriteRequest) (*model.Metrics, error) {
	chunk := make([]*mpb.Metric, 0)
	now := time.Now()

	for _, ts := range wr.Timeseries {
		attrs := map[string]string{}
		for _, l := range ts.Labels {
			attrs[l.Name] = l.Value
		}
		metricName := attrs[pmodel.MetricNameLabel]
		if metricName == "" {
			return nil, fmt.Errorf("metric name %q not found in attrs or empty", pmodel.MetricNameLabel)
		}
		delete(attrs, pmodel.MetricNameLabel)
		for _, s := range ts.Samples {
			dataPoints := make(map[string]*structpb.Value)
			if !math.IsNaN(s.Value) {
				dataPoints[metricName] = structpb.NewNumberValue(s.Value)
			}
			// converting to metric
			if len(dataPoints) > 0 {
				t := now
				if s.Timestamp > 0 {
					t = time.Unix(0, s.Timestamp*1000000)
				}
				m := &mpb.Metric{
					Name:         "prometheus_remote_write",
					TimeUnixNano: uint64(t.UnixNano()),
					Attributes:   attrs,
					DataPoints:   dataPoints,
				}
				chunk = append(chunk, m)
			}
		}
	}

	return &model.Metrics{Metrics: chunk}, nil
}
