package main

import (
	"context"
	"os"

	"github.com/recallsong/go-utils/logs"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/providers/metrics/common"
	"github.com/erda-project/erda/providers/metrics/report"
)

type define struct{}

func (d *define) Service() []string {
	return []string{"hello metric_report_client"}
}

func (d *define) Dependencies() []string {
	return []string{"metric-report-client"}
}

func (d *define) Description() string {
	return "hello for metric_report_client example"
}

type provider struct {
	Log        logs.Logger
	SendClient report.MetricReport
}

func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

func (p *provider) Init(ctx context.Context) error {
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	metric := []*common.Metric{
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
	}
	client := report.CreateReportClient(os.Getenv("ADDR"), os.Getenv("USERNAME"), os.Getenv("PASSWORD"))
	err := client.Send(metric)
	return err
}

func init() {
	servicehub.RegisterProvider("example", &define{})
}

func main() {
	hub := servicehub.New()
	hub.Run("examples", "", os.Args...)
}
