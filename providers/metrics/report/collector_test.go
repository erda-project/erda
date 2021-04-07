package report

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/erda-project/erda/providers/common"
)

func Test_reportClient_serialize(t *testing.T) {
	type fields struct {
		cfg        *config
		httpClient *http.Client
	}
	type args struct {
		group *NamedMetrics
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test_serialize",
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
				httpClient: new(http.Client),
			},
			args: args{
				group: &NamedMetrics{
					Name: "_metric_meta",
					Metrics: Metrics{
						&common.Metric{
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ReportClient{
				CFG:        tt.fields.cfg,
				HttpClient: tt.fields.httpClient,
			}
			got, err := c.serialize(tt.args.group)
			if (err != nil) != tt.wantErr {
				t.Errorf("serialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got)
		})
	}
}

func Test_reportClient_group(t *testing.T) {
	type fields struct {
		cfg        *config
		httpClient *http.Client
	}
	type args struct {
		in []*common.Metric
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test_group",
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
				httpClient: new(http.Client),
			},
			args: args{
				in: Metrics{
					&common.Metric{
						Name:      "span",
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
			c := &ReportClient{
				CFG:        tt.fields.cfg,
				HttpClient: tt.fields.httpClient,
			}
			if got := c.group(tt.args.in); got != nil {
				for _, v := range got {
					g := v
					fmt.Printf("%+v", *g)
				}
			}
		})
	}
}
