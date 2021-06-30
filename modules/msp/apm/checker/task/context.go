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

package task

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/erda-project/erda/modules/msp/apm/checker/plugins"
	"github.com/erda-project/erda/providers/metrics/report"
)

type reportContext struct {
	context.Context
	report report.MetricReport
}

func newTaskContext(ctx context.Context, r report.MetricReport) plugins.Context {
	if r == nil {
		return &writerContext{Context: ctx, w: os.Stdout}
	}
	return &reportContext{
		Context: ctx,
		report:  r,
	}
}

func (c *reportContext) Report(metrics ...*plugins.Metric) error {
	if len(metrics) > 0 {
		list := make(report.Metrics, len(metrics))
		now := time.Now().UnixNano()
		for i, m := range metrics {
			if m.Timestamp <= 0 {
				m.Timestamp = now
			}
			list[i] = &report.Metric{
				Name:      m.Name,
				Timestamp: m.Timestamp,
				Tags:      m.Tags,
				Fields:    m.Fields,
			}
		}
		return c.report.Send(list)
	}
	return nil
}

type writerContext struct {
	context.Context
	w io.Writer
}

func (c *writerContext) Report(metrics ...*plugins.Metric) error {
	now := time.Now().UnixNano()
	for _, m := range metrics {
		if m.Timestamp <= 0 {
			m.Timestamp = now
		}
		err := json.NewEncoder(c.w).Encode(m)
		if err != nil {
			return err
		}
	}
	return nil
}
