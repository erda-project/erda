package main

import (
	"context"
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/providers/metrics/query"
	"os"
	"time"
)

type define struct{}

func (d *define) Service() []string      { return []string{"hello"} }
func (d *define) Dependencies() []string { return []string{"metricq-client"} }
func (d *define) Description() string    { return "hello for example" }
func (d *define) Creator() servicehub.Creator {
	return func() servicehub.Provider {
		return &provider{}
	}
}

type provider struct {
	Log         logs.Logger
	QueryClient query.MetricQuery
}

func (p *provider) Init(ctx context.Context) error {
	return nil
}

func (p *provider) Run(ctx context.Context) error {
	request := query.CreateQueryRequest("docker_container_summary")
	now := time.Now()
	start, end := now.AddDate(0, 0, -1), now
	request = request.StartFrom(start).EndWith(end).Apply("avg", "fields.cpu_usage_percent")
	resp, err := p.QueryClient.QueryMetric(request)
	if err != nil {
		return err
	}
	// 获取单值数据对象
	point, err := resp.ReturnAsPoint()
	p.Log.Infof("metric_name : %s", point.Name)
	for _, data := range point.Data {
		p.Log.Infof("cpu_usage_percent : %f", data.Data)
	}
	return nil
}

func init() {
	servicehub.RegisterProvider("example", &define{})
}

func main() {
	hub := servicehub.New()
	hub.Run("examples", "", os.Args...)
}
