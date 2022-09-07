package metricmeta

import (
	"github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
)

type MetricMeta interface {
	GetMetricMetaByCache(scope, scopeID string, names ...string) ([]*pb.MetricMeta, error)
}
