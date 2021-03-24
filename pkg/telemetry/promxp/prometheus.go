package promxp

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erda-project/erda/pkg/telemetry/common"
)

func Start(addr string, name string) error {
	server := &http.Server{Addr: addr, Handler: Handler(name)}
	return server.ListenAndServe()
}

func Handler(name string) http.Handler {
	handler := promxpHandler()
	common.GetGlobalLabels()["metric_name"] = name
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Metric-Name", name)
		handler.ServeHTTP(resp, req)
	})
}

//func HealthyHandler(name string) http.Handler {
//	handler := health.HTTPHandler(name)
//	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
//		resp.Header().Set("Metric-Name", name)
//		handler.ServeHTTP(resp, req)
//	})
//}

func Register(cs ...prometheus.Collector) {
	MustRegister(cs...)
}

func normalizeKey(key string) string {
	key = strings.Replace(key, ".", "_", -1)
	key = strings.Replace(key, "-", "_", -1)
	return key
}

func getNamespace(names ...string) (namespace, subsystem string) {
	if len(names) > 0 {
		namespace = normalizeKey(names[0])
	}
	if len(names) > 1 {
		subsystem = normalizeKey(names[1])
	}
	return namespace, subsystem
}

// counter .
func NewCounter(field, description string, labels map[string]string, names ...string) prometheus.Counter {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: prometheus.Labels(labels),
	})
}

func NewCounterVec(field, description string, constLabels map[string]string, labelNames []string, names ...string) *prometheus.CounterVec {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: constLabels,
	}, labelNames)
}

func RegisterCounter(field, description string, labels map[string]string, names ...string) prometheus.Counter {
	counter := NewCounter(field, description, labels, names...)
	Register(counter)
	return counter
}

func RegisterCounterVec(field, description string, constLabels map[string]string, labelNames []string, names ...string) *prometheus.CounterVec {
	counter := NewCounterVec(field, description, constLabels, labelNames, names...)
	Register(counter)
	return counter
}

// gauge .
func NewGauge(field, description string, labels map[string]string, names ...string) prometheus.Gauge {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: prometheus.Labels(labels),
	})
}

func NewGaugeVec(field, description string, constLabels map[string]string, labelNames []string, names ...string) *prometheus.GaugeVec {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: constLabels,
	}, labelNames)
}

func RegisterGauge(field, description string, labels map[string]string, names ...string) prometheus.Gauge {
	gauge := NewGauge(field, description, labels, names...)
	Register(gauge)
	return gauge
}

func RegisterGaugeVec(field, description string, constLabels map[string]string, labelNames []string, names ...string) *prometheus.GaugeVec {
	gauge := NewGaugeVec(field, description, constLabels, labelNames, names...)
	Register(gauge)
	return gauge
}

// summary .
func NewSummary(field, description string, labels map[string]string, objectives map[float64]float64, names ...string) prometheus.Summary {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: prometheus.Labels(labels),
		Objectives:  objectives,
	})
}

func NewSummaryVec(field, description string, constLabels map[string]string, labelNames []string, objectives map[float64]float64, names ...string) *prometheus.SummaryVec {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: constLabels,
		Objectives:  objectives,
	}, labelNames)
}

func RegisterSummary(field, description string, labels map[string]string, objectives map[float64]float64, names ...string) prometheus.Summary {
	summary := NewSummary(field, description, labels, objectives, names...)
	Register(summary)
	return summary
}

func RegisterSummaryVec(field, description string, constLabels map[string]string, labelNames []string, objectives map[float64]float64, names ...string) *prometheus.SummaryVec {
	summary := NewSummaryVec(field, description, constLabels, labelNames, objectives, names...)
	Register(summary)
	return summary
}

// histogram ...
func NewHistogram(field, description string, labels map[string]string, buckets []float64, names ...string) prometheus.Histogram {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: prometheus.Labels(labels),
		Buckets:     buckets,
	})
}

func NewHistogramVec(field, description string, constLabels map[string]string, labelNames []string, buckets []float64, names ...string) *prometheus.HistogramVec {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: constLabels,
		Buckets:     buckets,
	}, labelNames)
}

func RegisterHistogram(field, description string, labels map[string]string, buckets []float64, names ...string) prometheus.Histogram {
	histogram := NewHistogram(field, description, labels, buckets, names...)
	Register(histogram)
	return histogram
}

func RegisterHistogramVec(field, description string, constLabels map[string]string, labelNames []string, buckets []float64, names ...string) *prometheus.HistogramVec {
	histogram := NewHistogramVec(field, description, constLabels, labelNames, buckets, names...)
	Register(histogram)
	return histogram
}

func emptyUntypeFunc() float64 {
	return 0
}

// untype .
func NewUntypedMetric(field, description string, labels map[string]string, value func() float64, names ...string) prometheus.UntypedFunc {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	if value == nil {
		value = emptyUntypeFunc
	}
	return prometheus.NewUntypedFunc(prometheus.UntypedOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: prometheus.Labels(labels),
	}, value)
}

func RegisterUntypedMetric(field, description string, labels map[string]string, value func() float64, names ...string) prometheus.UntypedFunc {
	metric := NewUntypedMetric(field, description, labels, value, names...)
	Register(metric)
	return metric
}

func NewMeter(field, description string, labels map[string]string, names ...string) Meter {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return newMeter(MeterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: prometheus.Labels(labels),
	}, EMPTY_LABLES)
}

func RegisterMeter(field, description string, labels map[string]string, names ...string) Meter {
	meter := NewMeter(field, description, labels, names...)
	Register(meter)
	return meter
}

func NewMeterVec(field, description string, constLabels map[string]string, labelNames []string, names ...string) *MeterVec {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return newMeterVec(MeterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: prometheus.Labels(constLabels),
	}, labelNames)
}

func RegisterMeterVec(field, description string, constLabels map[string]string, labelNames []string, names ...string) *MeterVec {
	meter := NewMeterVec(field, description, constLabels, labelNames, names...)
	Register(meter)
	return meter
}

// auto reset counter .
func NewAutoResetCounter(field, description string, labels map[string]string, names ...string) prometheus.Counter {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return newAutoResetCounter(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: prometheus.Labels(labels),
	}, EMPTY_LABLES)
}

func RegisterAutoResetCounter(field, description string, labels map[string]string, names ...string) prometheus.Counter {
	counter := NewAutoResetCounter(field, description, labels, names...)
	Register(counter)
	return counter
}

func NewAutoResetCounterVec(field, description string, constLabels map[string]string, labelNames []string, names ...string) *AutoResetCounterVec {
	field = normalizeKey(field)
	namespace, subsystem := getNamespace(names...)
	return newAutoResetCounterVec(prometheus.CounterOpts{
		Namespace:   namespace,
		Subsystem:   subsystem,
		Name:        field,
		Help:        description,
		ConstLabels: prometheus.Labels(constLabels),
	}, labelNames)
}

func RegisterAutoResetCounterVec(field, description string, constLabels map[string]string, labelNames []string, names ...string) *AutoResetCounterVec {
	counter := NewAutoResetCounterVec(field, description, constLabels, labelNames, names...)
	Register(counter)
	return counter
}
