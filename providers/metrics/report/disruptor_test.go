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

package report

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_disruptor_push(t *testing.T) {
	type fields struct {
		metrics  chan *Metric
		labels   GlobalLabel
		reporter Reporter
	}
	m := &Metric{
		Name:      "_metric_meta",
		Timestamp: 1614583470000,
		Tags: map[string]string{
			"cluster_name": "terminus-dev",
			"meta":         "true",
			"metric_name":  "application_db",
		},
		Fields: map[string]interface{}{
			"fields": []string{"value:number"},
			"tags":   []string{"is_edge", "org_id"},
		},
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test_test",
			fields: fields{
				metrics:  make(chan *Metric, 0),
				labels:   GlobalLabel{},
				reporter: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &disruptor{
				metrics:  tt.fields.metrics,
				labels:   tt.fields.labels,
				reporter: tt.fields.reporter,
				cfg: &config{
					ReportConfig: &ReportConfig{
						BufferSize: 100,
					},
				},
			}
			d.push()
			d.metrics <- m
		})
	}
}

func Test_disruptor_In(t *testing.T) {
	type fields struct {
		metrics  chan *Metric
		labels   GlobalLabel
		cfg      *config
		reporter Reporter
	}
	type args struct {
		metrics []*Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test_disruptor_in",
			fields: fields{
				cfg: &config{
					ReportConfig: &ReportConfig{
						Collector: &CollectorConfig{
							Addr:     "collector.default.svc.cluster.local:7076",
							UserName: "admin",
							Password: "Cqq",
							Retry:    2,
						},
					},
				},
				labels: map[string]string{
					"_meta":   "true",
					"_custom": "true",
				},
				metrics: make(chan *Metric),
			},
			args: args{
				metrics: []*Metric{
					{
						Name:      "_metric_meta",
						Timestamp: 1614583470000,
						Tags: map[string]string{
							"cluster_name": "terminus-dev",
							"meta":         "true",
							"metric_name":  "application_db",
						},
						Fields: map[string]interface{}{
							"fields": []string{"value:number"},
							"tags":   []string{"is_edge", "org_id"},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &disruptor{
				metrics:  tt.fields.metrics,
				labels:   tt.fields.labels,
				cfg:      tt.fields.cfg,
				reporter: tt.fields.reporter,
			}
			go func() {
				data := <-d.metrics
				fmt.Printf("the data is %+v", data)
			}()
			if err := d.In(tt.args.metrics...); (err != nil) != tt.wantErr {
				t.Errorf("In() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_disruptor_dataToMetric(t *testing.T) {
	type fields struct {
		metrics  chan *Metric
		labels   GlobalLabel
		cfg      *config
		reporter Reporter
	}
	type args struct {
		data []interface{}
	}
	metric := &Metric{
		Name:      "_metric_meta",
		Timestamp: 1614583470000,
		Tags: map[string]string{
			"cluster_name": "terminus-dev",
			"meta":         "true",
			"metric_name":  "application_db",
		},
		Fields: map[string]interface{}{
			"fields": []string{"value:number"},
			"tags":   []string{"is_edge", "org_id"},
		},
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []*Metric
	}{
		{
			name:   "test_dataToMetric",
			fields: fields{},
			args: args{
				data: []interface{}{
					metric,
				},
			},
			want: []*Metric{metric},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &disruptor{
				metrics:  tt.fields.metrics,
				labels:   tt.fields.labels,
				cfg:      tt.fields.cfg,
				reporter: tt.fields.reporter,
			}
			if got := d.dataToMetric(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("dataToMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}
