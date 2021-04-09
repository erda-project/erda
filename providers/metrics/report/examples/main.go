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

package main

import (
	"context"
	"os"

	"github.com/recallsong/go-utils/logs"

	"github.com/erda-project/erda-infra/base/servicehub"
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
	Log logs.Logger
	//SendClient report.MetricReport
	SendClient report.MetricReport
}

func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return func() servicehub.Provider {
			return &provider{}
		}
	}
}

func (p *provider) Init(ctx context.Context) error {
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	metric := []*report.Metric{
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
	client := p.SendClient.CreateReportClient(os.Getenv("COLLECTOR_ADDR"), os.Getenv("USERNAME"), os.Getenv("PASSWORD"))
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
