package report

import "github.com/erda-project/erda/providers/metrics/common"

type Reporter interface {
	Send(metrics []*common.Metric) error
}
