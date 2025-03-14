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
	"math"
	"testing"

	pmodel "github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"

	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
)

func Test_parseWriteRequest(t *testing.T) {
	type args struct {
		wr       *prompb.WriteRequest
		callback func(record *metric.Metric) error
	}
	ass := assert.New(t)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			args: args{
				wr: &prompb.WriteRequest{
					Timeseries: []*prompb.TimeSeries{
						{
							Labels: []*prompb.Label{
								{
									Name:  "n1",
									Value: "v1",
								},
								{
									Name:  pmodel.MetricNameLabel,
									Value: "cpu_user_total",
								},
								{
									Name:  pmodel.JobLabel,
									Value: "cpu",
								},
							},
							Samples: []*prompb.Sample{
								{
									Value:     1,
									Timestamp: 1658904849000,
								},
								{
									Value:     math.NaN(),
									Timestamp: 1658904849001,
								},
							},
						},
					},
				},
				callback: func(record *metric.Metric) error {
					ass.Equal(int64(1658904849000*1000000), record.Timestamp)
					ass.Equal("cpu", record.Name)
					ass.Equal(1, len(record.Fields))
					return nil
				},
			},
		},
		{
			name: "no name",
			args: args{
				wr: &prompb.WriteRequest{
					Timeseries: []*prompb.TimeSeries{
						{
							Labels: []*prompb.Label{
								{
									Name:  "n1",
									Value: "v1",
								},
								{
									Name:  pmodel.JobLabel,
									Value: "cpu",
								},
							},
							Samples: []*prompb.Sample{
								{
									Value:     1,
									Timestamp: 1658904849000,
								},
								{
									Value:     math.NaN(),
									Timestamp: 1658904849001,
								},
							},
						},
					},
				},
				callback: func(record *metric.Metric) error {
					ass.Equal(int64(1658904849000*1000000), record.Timestamp)
					ass.Equal("cpu", record.Name)
					ass.Equal(1, len(record.Fields))
					return nil
				},
			},
			wantErr: true,
		},
		{
			name: "no job",
			args: args{
				wr: &prompb.WriteRequest{
					Timeseries: []*prompb.TimeSeries{
						{
							Labels: []*prompb.Label{
								{
									Name:  "n1",
									Value: "v1",
								},
								{
									Name:  pmodel.MetricNameLabel,
									Value: "cpu_user_total",
								},
							},
							Samples: []*prompb.Sample{
								{
									Value:     1,
									Timestamp: 1658904849000,
								},
								{
									Value:     math.NaN(),
									Timestamp: 1658904849001,
								},
							},
						},
					},
				},
				callback: func(record *metric.Metric) error {
					ass.Equal(int64(1658904849000*1000000), record.Timestamp)
					ass.Equal("cpu", record.Name)
					ass.Equal(1, len(record.Fields))
					return nil
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseWriteRequest(tt.args.wr, nil, tt.args.callback); (err != nil) != tt.wantErr {
				t.Errorf("parseWriteRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
