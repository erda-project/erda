package report

import (
	"github.com/erda-project/erda/providers/metrics/common"
	"reflect"
	"testing"
)

func Test_buffer_Flush(t *testing.T) {
	type fields struct {
		count int
		max   int
		data  []*common.Metric
	}
	d := &common.Metric{
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
		want   []*common.Metric
	}{
		{
			name: "test_flush",
			fields: fields{
				count: 1,
				max:   1,
				data: Metrics{
					d,
				},
			},
			want: []*common.Metric{
				d,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &buffer{
				count: tt.fields.count,
				max:   tt.fields.max,
				data:  tt.fields.data,
			}
			if got := b.Flush(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Flush() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buffer_Add(t *testing.T) {
	type fields struct {
		count int
		max   int
		data  []*common.Metric
	}
	type args struct {
		v *common.Metric
	}
	d := &common.Metric{
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
		want   bool
	}{
		{
			name: "test_add",
			fields: fields{
				count: 1,
				max:   2,
				data: []*common.Metric{
					d,
					nil,
				},
			},
			args: args{
				d,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &buffer{
				count: tt.fields.count,
				max:   tt.fields.max,
				data:  tt.fields.data,
			}
			if got := b.Add(tt.args.v); got != tt.want {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}
