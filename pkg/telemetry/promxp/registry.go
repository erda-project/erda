package promxp

import "github.com/prometheus/client_golang/prometheus"

var (
	promxpRegistry                         = prometheus.NewRegistry()
	promxpRegisterer prometheus.Registerer = promxpRegistry
	promxpGatherer   prometheus.Gatherer   = promxpRegistry
)

func MustRegister(cs ...prometheus.Collector) {
	promxpRegisterer.MustRegister(cs...)
}
