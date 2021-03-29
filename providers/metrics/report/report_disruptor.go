package report

import (
	"github.com/erda-project/erda/providers/metrics/common"
	report_config "github.com/erda-project/erda/providers/metrics/config"
	"time"
)

type Disruptor interface {
	In(metrics ...*common.Metric) error
}

type disruptor struct {
	metrics chan *common.Metric
	labels  common.GlobalLabel

	reporter Reporter
}

func (d *disruptor) In(metrics ...*common.Metric) error {
	if len(metrics) > 0 {
		for _, metric := range metrics {
			for k, v := range d.labels {
				if _, ok := metric.Tags[k]; !ok {
					metric.Tags[k] = v
				}
			}
			d.metrics <- metric
		}
	}
	return nil
}

func (d *disruptor) push() {
	go func(queue chan *common.Metric, reporter Reporter, queueSize int) {
		buf := newBuffer(queueSize)
		ticker := time.NewTicker(time.Second * time.Duration(5))
		for {
			select {
			case metric, ok := <-queue:
				if !ok {
					if metrics := buf.Flush(); len(metrics) != 0 {
						_ = reporter.Send(metrics)
					}
					break
				}
				buf.Add(metric)
				if buf.IsOverFlow() {
					_ = reporter.Send(buf.Flush())
				}
			case <-ticker.C:
				if !buf.IsEmpty() {
					if metrics := buf.Flush(); len(metrics) != 0 {
						_ = reporter.Send(metrics)
					}
				}
			}
		}
	}(d.metrics, d.reporter, report_config.GlobalConfig().ReportConfig.BufferSize)
}
