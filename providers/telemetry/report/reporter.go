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
	"github.com/erda-project/erda/providers/telemetry/common"
	c "github.com/erda-project/erda/providers/telemetry/config"
	"github.com/pkg/errors"
)

type Reporter interface {
	Send(metrics []*common.Metric) error
}

var NoopReporter = &noopReporter{}

type noopReporter struct {
}

func (r *noopReporter) Send([]*common.Metric) error {
	return nil
}

func createReporter(cfg *c.ReportConfig) Reporter {
	var (
		reporter Reporter
		err      error
	)
	if cfg.Mode == c.STRICT_MODE {
		reporter, err = newCollectorReporter(cfg.Collector)
	} else if cfg.Mode == c.PERFORMANCE_MODE {
		reporter, err = newTelegrafReporter(cfg.UdpHost, cfg.UdpPort)
	} else {
		err = errors.New("invalid report mode")
	}
	if err != nil {
		return NoopReporter
	}
	return reporter
}