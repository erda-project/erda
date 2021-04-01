package report

import (
	"testing"

	"github.com/erda-project/erda/providers/metrics/common"
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
				reporter: NoopReporter,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &disruptor{
				metrics:  tt.fields.metrics,
				labels:   tt.fields.labels,
				reporter: tt.fields.reporter,
			}
			d.push()
			d.metrics <- m
		})
	}
}
