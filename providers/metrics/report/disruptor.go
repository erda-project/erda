package report

import (
	"time"

	parallel "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda/providers/metrics/common"
)

type Disruptor interface {
	In(metrics ...*common.Metric) error
}

type disruptor struct {
	metrics  chan *common.Metric
	labels   common.GlobalLabel
	cfg      *config
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

func (d *disruptor) dataToMetric(data []interface{}) []*common.Metric {
	resultArr := make([]*common.Metric, 0)
	for _, v := range data {
		m, ok := v.(*common.Metric)
		if ok {
			resultArr = append(resultArr, m)
		}
	}
	return resultArr
}

func (d *disruptor) push() {
	go func(queue chan *common.Metric, reporter Reporter, queueSize int) {
		reportWrite := NewReportWrite(queueSize)
		buf := parallel.NewBuffer(reportWrite, queueSize)
		ticker := time.NewTicker(time.Second * time.Duration(5))
		for {
			select {
			case metric, ok := <-queue:
				if !ok {
					if res, err := buf.WriteN(metric); res != 0 && err == nil {
						data := buf.Data()
						resultArr := d.dataToMetric(data[:res])
						if resultArr != nil {
							_ = reporter.Send(resultArr)
						}
					}
					break
				}
				buf.Write(metric)
			case <-ticker.C:
				if data := buf.Data(); len(data) > 0 {
					buf.Flush()
					resultArr := d.dataToMetric(data[:])
					if resultArr != nil {
						_ = reporter.Send(resultArr)
					}
				}
			}
		}
	}(d.metrics, d.reporter, d.cfg.ReportConfig.BufferSize)
}
