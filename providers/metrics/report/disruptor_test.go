package report

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/erda-project/erda/providers/common"
)

func Test_disruptor_push(t *testing.T) {
	type fields struct {
		metrics  chan *common.Metric
		labels   common.GlobalLabel
		reporter Reporter
	}
	m := &common.Metric{
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
				metrics:  make(chan *common.Metric, 0),
				labels:   common.GlobalLabel{},
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
		metrics  chan *common.Metric
		labels   common.GlobalLabel
		cfg      *config
		reporter Reporter
	}
	type args struct {
		metrics []*common.Metric
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
				metrics: make(chan *common.Metric),
			},
			args: args{
				metrics: []*common.Metric{
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
		metrics  chan *common.Metric
		labels   common.GlobalLabel
		cfg      *config
		reporter Reporter
	}
	type args struct {
		data []interface{}
	}
	metric := &common.Metric{
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
		want   []*common.Metric
	}{
		{
			name:   "test_dataToMetric",
			fields: fields{},
			args: args{
				data: []interface{}{
					metric,
				},
			},
			want: []*common.Metric{metric},
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
