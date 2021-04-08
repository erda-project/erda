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

	"github.com/erda-project/erda/providers/telemetry/common"
	c "github.com/erda-project/erda/providers/telemetry/config"
)

var (
	DefaultDisruptor = NewDisruptor(nil)
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
	}(d.metrics, d.reporter, c.GlobalConfig().ReportConfig.BufferSize)
}

func NewDisruptor(cfg *c.ReportConfig) Disruptor {
	if cfg == nil {
		cfg = c.GlobalConfig().ReportConfig
	}
	d := &disruptor{
		metrics:  make(chan *common.Metric, cfg.BufferSize+64),
		labels:   common.GetGlobalLabels(),
		reporter: createReporter(cfg),
	}
	d.push()
	return d
}
