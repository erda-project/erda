package report

import "github.com/erda-project/erda/providers/common"

type Reporter interface {
	Send(metrics []*common.Metric) error
}
