package promxp

import (
	"errors"
	"math"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

type autoResetCounter struct {
	self prometheus.Metric
	desc *prometheus.Desc

	bitsValue uint64
	intValue  uint64

	labelPairs []*dto.LabelPair

	now func() time.Time
}

type AutoResetCounterVec struct {
	*metricVec
}

func (arc *autoResetCounter) Desc() *prometheus.Desc {
	return arc.desc
}

// Describe implements Collector.
func (arc *autoResetCounter) Describe(ch chan<- *prometheus.Desc) {
	ch <- arc.self.Desc()
}

// Collect implements Collector.
func (arc *autoResetCounter) Collect(ch chan<- prometheus.Metric) {
	ch <- arc.self
}

func (arc *autoResetCounter) Add(v float64) {
	if v < 0 {
		panic(errors.New("counter cannot decrease in value"))
	}

	uintValue := uint64(v)
	if float64(uintValue) == v {
		atomic.AddUint64(&arc.intValue, uintValue)
		return
	}

	for {
		oldBits := atomic.LoadUint64(&arc.bitsValue)
		newBits := math.Float64bits(math.Float64frombits(oldBits) + v)
		if atomic.CompareAndSwapUint64(&arc.bitsValue, oldBits, newBits) {
			return
		}
	}
}

func (arc *autoResetCounter) Inc() {
	atomic.AddUint64(&arc.intValue, 1)
}

func (arc *autoResetCounter) Write(out *dto.Metric) error {
	floatValue := math.Float64frombits(atomic.LoadUint64(&arc.bitsValue))
	uintValue := atomic.LoadUint64(&arc.intValue)
	currentValue := floatValue + float64(uintValue)

	return arc.populateMetric(currentValue, arc.labelPairs, out)
}

func (arc *autoResetCounter) populateMetric(v float64, labelPairs []*dto.LabelPair, m *dto.Metric) error {
	m.Label = labelPairs
	m.Counter = &dto.Counter{Value: proto.Float64(v)}
	arc.reset()
	return nil
}

// init provides the selfCollector with a reference to the metric it is supposed
// to collect. It is usually called within the factory function to create a
// metric. See example.
func (arc *autoResetCounter) init(self prometheus.Metric) {
	arc.self = self
}

func (arc *autoResetCounter) reset() {
	atomic.StoreUint64(&arc.bitsValue, math.Float64bits(0))
	atomic.StoreUint64(&arc.intValue, 0)
}

func newAutoResetCounter(opts prometheus.CounterOpts, variableLabels []string, labelValues ...string) prometheus.Counter {
	if opts.ConstLabels == nil {
		opts.ConstLabels = map[string]string{}
	}

	desc := prometheus.NewDesc(
		prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		nil,
		opts.ConstLabels,
	)
	result := &autoResetCounter{desc: desc, labelPairs: makeLabelPairs(opts.ConstLabels, variableLabels, labelValues), now: time.Now}
	// Init self-collection.
	result.init(result)
	return result
}

func (v *AutoResetCounterVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Counter, error) {
	metric, err := v.metricVec.GetMetricWithLabelValues(lvs...)
	if metric != nil {
		return metric.(prometheus.Counter), err
	}
	return nil, err
}

func (v *AutoResetCounterVec) WithLabelValues(lvs ...string) prometheus.Counter {
	c, err := v.GetMetricWithLabelValues(lvs...)
	if err != nil {
		panic(err)
	}
	return c
}

func newAutoResetCounterVec(opts prometheus.CounterOpts, variableLabels []string) *AutoResetCounterVec {
	fqName := prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name)
	desc := prometheus.NewDesc(
		fqName,
		opts.Help,
		variableLabels,
		opts.ConstLabels,
	)
	return &AutoResetCounterVec{
		metricVec: newMetricVec(desc, variableLabels, func(lvs ...string) prometheus.Metric {
			if len(lvs) != len(variableLabels) {
				panic(makeInconsistentCardinalityError(fqName, variableLabels, lvs))
			}
			return newAutoResetCounter(opts, variableLabels, lvs...)
		}),
	}
}
