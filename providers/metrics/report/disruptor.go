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
	"time"

	parallel "github.com/erda-project/erda-infra/pkg/parallel-writer"
)

type Disruptor interface {
	In(metrics ...*Metric) error
}

type disruptor struct {
	metrics  chan *Metric
	labels   GlobalLabel
	cfg      *config
	reporter Reporter
}

func (d *disruptor) In(metrics ...*Metric) error {
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

func (d *disruptor) dataToMetric(data []interface{}) []*Metric {
	resultArr := make([]*Metric, 0)
	for _, v := range data {
		m, ok := v.(*Metric)
		if ok {
			resultArr = append(resultArr, m)
		}
	}
	return resultArr
}

func (d *disruptor) push() {
	go func(queue chan *Metric, reporter Reporter, queueSize int) {
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
