package promxp

import (
	"github.com/erda-project/erda/pkg/telemetry/common"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	HTTPPathMetrics       = "/metrics"
	HTTPPathMetricsHealth = "/metrics/health"
)

func newComponentCollector() prometheus.Collector {
	return NewUntypedMetric("_global_info", "component information", common.GetGlobalLabels(), nil)
}

func init() {
	MustRegister(newComponentCollector())
}
