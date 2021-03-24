package promxp

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

var DefBucket uint32 = 60

type Meter interface {
	prometheus.Metric
	prometheus.Collector

	Mark(int64)
	Snapshot() *MeterSnapshot
}

type MeterOpts struct {

	// Namespace, Subsystem, and Name are components of the fully-qualified
	// name of the Summary (created by joining these components with
	// "_"). Only Name is mandatory, the others merely help structuring the
	// name. Note that the fully-qualified name of the Summary must be a
	// valid Prometheus metric name.
	Namespace string
	Subsystem string
	Name      string

	// Help provides information about this Summary.
	//
	// Metrics with the same fully-qualified name must have the same Help
	// string.
	Help string

	// ConstLabels are used to attach fixed labels to this metric. Metrics
	// with the same fully-qualified name must have the same label names in
	// their ConstLabels.
	//
	// Due to the way a Summary is represented in the Prometheus text format
	// and how it is handled by the Prometheus server internally, “quantile”
	// is an illegal label name. Construction of a Summary or SummaryVec
	// will panic if this label name is used in ConstLabels.
	//
	// ConstLabels are only used rarely. In particular, do not use them to
	// attach the same labels to all your metrics. Those use cases are
	// better covered by target labels set by the scraping Prometheus
	// server, or by one specific metric (e.g. a build_info or a
	// machine_role metric). See also
	// https://prometheus.io/docs/instrumenting/writing_exporters/#target-labels,-not-static-scraped-labels
	ConstLabels prometheus.Labels

	Bucket uint32
}

// MeterSnapshot is a read-only copy of another Meter.
type MeterSnapshot struct {
	count int64
	rate  uint64
}

func (m *MeterSnapshot) Count() int64 { return m.count }

func (m *MeterSnapshot) Rate() float64 { return math.Float64frombits(m.rate) }

// Snapshot returns the snapshot.
//func (m *meterSnapshot) Snapshot() meterSnapshot { return m }

type meter struct {
	self prometheus.Metric
	desc *prometheus.Desc

	labelPairs []*dto.LabelPair

	snapshot *MeterSnapshot
	bucket   uint32
	ewma     EWMA
	ticker   *time.Ticker
}

func (m *meter) Desc() *prometheus.Desc {
	return m.desc
}

// Describe implements Collector.
func (m *meter) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.self.Desc()
}

// Collect implements Collector.
func (m *meter) Collect(ch chan<- prometheus.Metric) {
	ch <- m.self
}

func (m *meter) Write(out *dto.Metric) error {
	return m.populateMetric(math.Float64frombits(atomic.LoadUint64(&m.snapshot.rate)), m.labelPairs, out)
}

// init provides the selfCollector with a reference to the metric it is supposed
// to collect. It is usually called within the factory function to create a
// metric. See example.
func (m *meter) init(self prometheus.Metric) {
	m.self = self
	min := float64(m.bucket / 60)
	m.ewma = NewEWMA(1 - math.Exp(-5.0/60.0/min))
	go m.tick()
}

func (m *meter) Mark(v int64) {
	atomic.AddInt64(&m.snapshot.count, v)
	m.ewma.Update(v)
	m.updateSnapshot()
}

func (m *meter) Snapshot() *MeterSnapshot {
	copiedSnapshot := MeterSnapshot{
		count: atomic.LoadInt64(&m.snapshot.count),
		rate:  atomic.LoadUint64(&m.snapshot.rate),
	}
	return &copiedSnapshot
}

func (m *meter) updateSnapshot() {
	rate := math.Float64bits(m.ewma.Rate())
	atomic.StoreUint64(&m.snapshot.rate, rate)
}

func (m *meter) tick() {
	for {
		select {
		case <-m.ticker.C:
			m.ewma.Tick()
			m.updateSnapshot()
		}
	}
}

func (m *meter) populateMetric(v float64, labelPairs []*dto.LabelPair, metric *dto.Metric) error {
	labels := append(labelPairs, &dto.LabelPair{
		Name:  proto.String("_custom_prome_type"),
		Value: proto.String("meter"),
	})
	metric.Label = labels
	metric.Untyped = &dto.Untyped{Value: proto.Float64(v)}
	return nil
}

func newMeter(opts MeterOpts, variableLabels []string, labelValues ...string) Meter {
	if opts.ConstLabels == nil {
		opts.ConstLabels = map[string]string{}
	}
	if opts.Bucket <= 0 {
		opts.Bucket = DefBucket
	}

	desc := prometheus.NewDesc(
		prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name),
		opts.Help,
		nil,
		opts.ConstLabels,
	)

	result := &meter{
		desc:       desc,
		labelPairs: makeLabelPairs(opts.ConstLabels, variableLabels, labelValues),
		bucket:     opts.Bucket,
		snapshot:   &MeterSnapshot{},
		ticker:     time.NewTicker(5e9),
	}
	result.init(result)
	return result
}

type MeterVec struct {
	*metricVec
}

func (v *MeterVec) GetMetricWithLabelValues(lvs ...string) (Meter, error) {
	metric, err := v.metricVec.GetMetricWithLabelValues(lvs...)
	if metric != nil {
		return metric.(Meter), err
	}
	return nil, err
}

func (v *MeterVec) WithLabelValues(lvs ...string) Meter {
	c, err := v.GetMetricWithLabelValues(lvs...)
	if err != nil {
		panic(err)
	}
	return c
}

// NewCounterVec creates a new CounterVec based on the provided CounterOpts and
// partitioned by the given label names.
func newMeterVec(opts MeterOpts, variableLabels []string) *MeterVec {
	fqName := prometheus.BuildFQName(opts.Namespace, opts.Subsystem, opts.Name)
	desc := prometheus.NewDesc(
		fqName,
		opts.Help,
		variableLabels,
		opts.ConstLabels,
	)
	return &MeterVec{
		metricVec: newMetricVec(desc, variableLabels, func(lvs ...string) prometheus.Metric {
			if len(lvs) != len(variableLabels) {
				panic(makeInconsistentCardinalityError(fqName, variableLabels, lvs))
			}
			return newMeter(opts, variableLabels, lvs...)
		}),
	}
}
