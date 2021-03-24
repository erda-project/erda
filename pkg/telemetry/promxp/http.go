package promxp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func promxpHandler() http.Handler {
	return promhttp.InstrumentMetricHandler(
		promxpRegisterer, promhttp.HandlerFor(promxpGatherer, promhttp.HandlerOpts{}),
	)
}
