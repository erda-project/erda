package report

import (
	"errors"
	"github.com/erda-project/erda/pkg/telemetry/common"
	"github.com/erda-project/erda/pkg/telemetry/config"
)

type Reporter interface {
	Send(metrics []*common.Metric) error
}

type noopReporter struct {
}

var NoopReporter = &noopReporter{}

func (r *noopReporter) Send([]*common.Metric) error {
	return nil
}

func createReporter(cfg *config.ReportConfig) Reporter {
	var (
		reporter Reporter
		err      error
	)
	if cfg.Mode == config.STRICT_MODE {
		reporter, err = newCollectorReporter(cfg.Collector)
	} else if cfg.Mode == config.PERFORMANCE_MODE {
		reporter, err = newTelegrafReporter(cfg.UdpHost, cfg.UdpPort)
	} else {
		err = errors.New("invalid report mode")
	}
	if err != nil {
		return NoopReporter
	}
	return reporter
}
