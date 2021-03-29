package report

import "github.com/erda-project/erda/providers/metrics/common"

type Reporter interface {
	Send(metrics []*common.Metric) error
}

type noopReporter struct {
}

var NoopReporter = &noopReporter{}

func (r *noopReporter) Send([]*common.Metric) error {
	return nil
}
