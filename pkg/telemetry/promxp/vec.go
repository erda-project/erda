// from https://github.com/prometheus/client_golang/blob/master/prometheus/vec.go

package promxp

import (
	"errors"
	"fmt"
	"sync"
	"unicode/utf8"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
)

// Inline and byte-free variant of hash/fnv's fnv64a.

const (
	offset64 = 14695981039346656037
	prime64  = 1099511628211
)

// hashNew initializies a new fnv64a hash value.
func hashNew() uint64 {
	return offset64
}

// hashAdd adds a string to a fnv64a hash value, returning the updated hash.
func hashAdd(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}
	return h
}

// hashAddByte adds a byte to a fnv64a hash value, returning the updated hash.
func hashAddByte(h uint64, b byte) uint64 {
	h ^= uint64(b)
	h *= prime64
	return h
}

//type variableLabel []string

// metricVec is a Collector to bundle metrics of the same name that differ in
// their label values. metricVec is not used directly (and therefore
// unexported). It is used as a building block for implementations of vectors of
// a given metric type, like GaugeVec, CounterVec, SummaryVec, and HistogramVec.
// It also handles label currying.
type metricVec struct {
	*metricMap

	curry []curriedLabelValue

	// hashAdd and hashAddByte can be replaced for testing collision handling.
	hashAdd     func(h uint64, s string) uint64
	hashAddByte func(h uint64, b byte) uint64
}

// newMetricVec returns an initialized metricVec.
func newMetricVec(desc *prometheus.Desc, variableLabels []string, newMetric func(lvs ...string) prometheus.Metric) *metricVec {
	return &metricVec{
		metricMap: &metricMap{
			metrics:        map[uint64][]metricWithLabelValues{},
			desc:           desc,
			variableLabels: variableLabels,
			newMetric:      newMetric,
		},
		hashAdd:     hashAdd,
		hashAddByte: hashAddByte,
	}
}

// DeleteLabelValues removes the metric where the variable labels are the same
// as those passed in as labels (same order as the VariableLabels in Desc). It
// returns true if a metric was deleted.
//
// It is not an error if the number of label values is not the same as the
// number of VariableLabels in Desc. However, such inconsistent label count can
// never match an actual metric, so the method will always return false in that
// case.
//
// Note that for more than one label value, this method is prone to mistakes
// caused by an incorrect order of arguments. Consider Delete(Labels) as an
// alternative to avoid that type of mistake. For higher label numbers, the
// latter has a much more readable (albeit more verbose) syntax, but it comes
// with a performance overhead (for creating and processing the Labels map).
// See also the CounterVec example.
func (m *metricVec) DeleteLabelValues(lvs ...string) bool {
	h, err := m.hashLabelValues(lvs)
	if err != nil {
		return false
	}

	return m.metricMap.deleteByHashWithLabelValues(h, lvs, m.curry)
}

// Delete deletes the metric where the variable labels are the same as those
// passed in as labels. It returns true if a metric was deleted.
//
// It is not an error if the number and names of the Labels are inconsistent
// with those of the VariableLabels in Desc. However, such inconsistent Labels
// can never match an actual metric, so the method will always return false in
// that case.
//
// This method is used for the same purpose as DeleteLabelValues(...string). See
// there for pros and cons of the two methods.
func (m *metricVec) Delete(labels prometheus.Labels) bool {
	h, err := m.hashLabels(labels)
	if err != nil {
		return false
	}

	return m.metricMap.deleteByHashWithLabels(h, labels, m.curry)
}

// Without explicit forwarding of Describe, Collect, Reset, those methods won't
// show up in GoDoc.

// Describe implements Collector.
func (m *metricVec) Describe(ch chan<- *prometheus.Desc) { m.metricMap.Describe(ch) }

// Collect implements Collector.
func (m *metricVec) Collect(ch chan<- prometheus.Metric) { m.metricMap.Collect(ch) }

// Reset deletes all metrics in this vector.
func (m *metricVec) Reset() { m.metricMap.Reset() }

func (m *metricVec) CurryWith(labels prometheus.Labels) (*metricVec, error) {
	var (
		newCurry []curriedLabelValue
		oldCurry = m.curry
		iCurry   int
	)
	for i, label := range m.variableLabels {
		val, ok := labels[label]
		if iCurry < len(oldCurry) && oldCurry[iCurry].index == i {
			if ok {
				return nil, fmt.Errorf("label name %q is already curried", label)
			}
			newCurry = append(newCurry, oldCurry[iCurry])
			iCurry++
		} else {
			if !ok {
				continue // Label stays uncurried.
			}
			newCurry = append(newCurry, curriedLabelValue{i, val})
		}
	}
	if l := len(oldCurry) + len(labels) - len(newCurry); l > 0 {
		return nil, fmt.Errorf("%d unknown label(s) found during currying", l)
	}

	return &metricVec{
		metricMap:   m.metricMap,
		curry:       newCurry,
		hashAdd:     m.hashAdd,
		hashAddByte: m.hashAddByte,
	}, nil
}

func (m *metricVec) GetMetricWithLabelValues(lvs ...string) (prometheus.Metric, error) {
	h, err := m.hashLabelValues(lvs)
	if err != nil {
		return nil, err
	}

	return m.metricMap.getOrCreateMetricWithLabelValues(h, lvs, m.curry), nil
}

func (m *metricVec) GetMetricWith(labels prometheus.Labels) (prometheus.Metric, error) {
	h, err := m.hashLabels(labels)
	if err != nil {
		return nil, err
	}

	return m.metricMap.getOrCreateMetricWithLabels(h, labels, m.curry), nil
}

func (m *metricVec) hashLabelValues(vals []string) (uint64, error) {
	if err := validateLabelValues(vals, len(m.variableLabels)-len(m.curry)); err != nil {
		return 0, err
	}

	var (
		h             = hashNew()
		curry         = m.curry
		iVals, iCurry int
	)
	for i := 0; i < len(m.variableLabels); i++ {
		if iCurry < len(curry) && curry[iCurry].index == i {
			h = m.hashAdd(h, curry[iCurry].value)
			iCurry++
		} else {
			h = m.hashAdd(h, vals[iVals])
			iVals++
		}
		h = m.hashAddByte(h, model.SeparatorByte)
	}
	return h, nil
}

func (m *metricVec) hashLabels(labels prometheus.Labels) (uint64, error) {
	if err := validateValuesInLabels(labels, len(m.variableLabels)-len(m.curry)); err != nil {
		return 0, err
	}

	var (
		h      = hashNew()
		curry  = m.curry
		iCurry int
	)
	for i, label := range m.variableLabels {
		val, ok := labels[label]
		if iCurry < len(curry) && curry[iCurry].index == i {
			if ok {
				return 0, fmt.Errorf("label name %q is already curried", label)
			}
			h = m.hashAdd(h, curry[iCurry].value)
			iCurry++
		} else {
			if !ok {
				return 0, fmt.Errorf("label name %q missing in label map", label)
			}
			h = m.hashAdd(h, val)
		}
		h = m.hashAddByte(h, model.SeparatorByte)
	}
	return h, nil
}

// metricWithLabelValues provides the metric and its label values for
// disambiguation on hash collision.
type metricWithLabelValues struct {
	values []string
	metric prometheus.Metric
}

// curriedLabelValue sets the curried value for a label at the given index.
type curriedLabelValue struct {
	index int
	value string
}

// metricMap is a helper for metricVec and shared between differently curried
// metricVecs.
type metricMap struct {
	mtx            sync.RWMutex // Protects metrics.
	metrics        map[uint64][]metricWithLabelValues
	desc           *prometheus.Desc
	variableLabels []string
	newMetric      func(labelValues ...string) prometheus.Metric
}

// Describe implements Collector. It will send exactly one Desc to the provided
// channel.
func (m *metricMap) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.desc
}

// Collect implements Collector.
func (m *metricMap) Collect(ch chan<- prometheus.Metric) {
	m.mtx.RLock()
	defer m.mtx.RUnlock()

	for _, metrics := range m.metrics {
		for _, metric := range metrics {
			ch <- metric.metric
		}
	}
}

// Reset deletes all metrics in this vector.
func (m *metricMap) Reset() {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	for h := range m.metrics {
		delete(m.metrics, h)
	}
}

// deleteByHashWithLabelValues removes the metric from the hash bucket h. If
// there are multiple matches in the bucket, use lvs to select a metric and
// remove only that metric.
func (m *metricMap) deleteByHashWithLabelValues(
	h uint64, lvs []string, curry []curriedLabelValue,
) bool {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	metrics, ok := m.metrics[h]
	if !ok {
		return false
	}

	i := findMetricWithLabelValues(metrics, lvs, curry)
	if i >= len(metrics) {
		return false
	}

	if len(metrics) > 1 {
		m.metrics[h] = append(metrics[:i], metrics[i+1:]...)
	} else {
		delete(m.metrics, h)
	}
	return true
}

// deleteByHashWithLabels removes the metric from the hash bucket h. If there
// are multiple matches in the bucket, use lvs to select a metric and remove
// only that metric.
func (m *metricMap) deleteByHashWithLabels(
	h uint64, labels prometheus.Labels, curry []curriedLabelValue,
) bool {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	metrics, ok := m.metrics[h]
	if !ok {
		return false
	}
	i := findMetricWithLabels(m.variableLabels, metrics, labels, curry)
	if i >= len(metrics) {
		return false
	}

	if len(metrics) > 1 {
		m.metrics[h] = append(metrics[:i], metrics[i+1:]...)
	} else {
		delete(m.metrics, h)
	}
	return true
}

// getOrCreateMetricWithLabelValues retrieves the metric by hash and label value
// or creates it and returns the new one.
//
// This function holds the mutex.
func (m *metricMap) getOrCreateMetricWithLabelValues(
	hash uint64, lvs []string, curry []curriedLabelValue,
) prometheus.Metric {
	m.mtx.RLock()
	metric, ok := m.getMetricWithHashAndLabelValues(hash, lvs, curry)
	m.mtx.RUnlock()
	if ok {
		return metric
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()
	metric, ok = m.getMetricWithHashAndLabelValues(hash, lvs, curry)
	if !ok {
		inlinedLVs := inlineLabelValues(lvs, curry)
		metric = m.newMetric(inlinedLVs...)
		m.metrics[hash] = append(m.metrics[hash], metricWithLabelValues{values: inlinedLVs, metric: metric})
	}
	return metric
}

// getOrCreateMetricWithLabelValues retrieves the metric by hash and label value
// or creates it and returns the new one.
//
// This function holds the mutex.
func (m *metricMap) getOrCreateMetricWithLabels(
	hash uint64, labels prometheus.Labels, curry []curriedLabelValue,
) prometheus.Metric {
	m.mtx.RLock()
	metric, ok := m.getMetricWithHashAndLabels(hash, labels, curry)
	m.mtx.RUnlock()
	if ok {
		return metric
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()
	metric, ok = m.getMetricWithHashAndLabels(hash, labels, curry)
	if !ok {
		lvs := extractLabelValues(m.variableLabels, labels, curry)
		metric = m.newMetric(lvs...)
		m.metrics[hash] = append(m.metrics[hash], metricWithLabelValues{values: lvs, metric: metric})
	}
	return metric
}

// getMetricWithHashAndLabelValues gets a metric while handling possible
// collisions in the hash space. Must be called while holding the read mutex.
func (m *metricMap) getMetricWithHashAndLabelValues(
	h uint64, lvs []string, curry []curriedLabelValue,
) (prometheus.Metric, bool) {
	metrics, ok := m.metrics[h]
	if ok {
		if i := findMetricWithLabelValues(metrics, lvs, curry); i < len(metrics) {
			return metrics[i].metric, true
		}
	}
	return nil, false
}

// getMetricWithHashAndLabels gets a metric while handling possible collisions in
// the hash space. Must be called while holding read mutex.
func (m *metricMap) getMetricWithHashAndLabels(
	h uint64, labels prometheus.Labels, curry []curriedLabelValue,
) (prometheus.Metric, bool) {
	metrics, ok := m.metrics[h]
	if ok {
		if i := findMetricWithLabels(m.variableLabels, metrics, labels, curry); i < len(metrics) {
			return metrics[i].metric, true
		}
	}
	return nil, false
}

// findMetricWithLabelValues returns the index of the matching metric or
// len(metrics) if not found.
func findMetricWithLabelValues(
	metrics []metricWithLabelValues, lvs []string, curry []curriedLabelValue,
) int {
	for i, metric := range metrics {
		if matchLabelValues(metric.values, lvs, curry) {
			return i
		}
	}
	return len(metrics)
}

// findMetricWithLabels returns the index of the matching metric or len(metrics)
// if not found.
func findMetricWithLabels(
	variableLabels []string, metrics []metricWithLabelValues, labels prometheus.Labels, curry []curriedLabelValue,
) int {
	for i, metric := range metrics {
		if matchLabels(variableLabels, metric.values, labels, curry) {
			return i
		}
	}
	return len(metrics)
}

func matchLabelValues(values []string, lvs []string, curry []curriedLabelValue) bool {
	if len(values) != len(lvs)+len(curry) {
		return false
	}
	var iLVs, iCurry int
	for i, v := range values {
		if iCurry < len(curry) && curry[iCurry].index == i {
			if v != curry[iCurry].value {
				return false
			}
			iCurry++
			continue
		}
		if v != lvs[iLVs] {
			return false
		}
		iLVs++
	}
	return true
}

func matchLabels(variableLabels []string, values []string, labels prometheus.Labels, curry []curriedLabelValue) bool {
	if len(values) != len(labels)+len(curry) {
		return false
	}
	iCurry := 0
	for i, k := range variableLabels {
		if iCurry < len(curry) && curry[iCurry].index == i {
			if values[i] != curry[iCurry].value {
				return false
			}
			iCurry++
			continue
		}
		if values[i] != labels[k] {
			return false
		}
	}
	return true
}

func extractLabelValues(variableLabels []string, labels prometheus.Labels, curry []curriedLabelValue) []string {
	labelValues := make([]string, len(labels)+len(curry))
	iCurry := 0
	for i, k := range variableLabels {
		if iCurry < len(curry) && curry[iCurry].index == i {
			labelValues[i] = curry[iCurry].value
			iCurry++
			continue
		}
		labelValues[i] = labels[k]
	}
	return labelValues
}

func inlineLabelValues(lvs []string, curry []curriedLabelValue) []string {
	labelValues := make([]string, len(lvs)+len(curry))
	var iCurry, iLVs int
	for i := range labelValues {
		if iCurry < len(curry) && curry[iCurry].index == i {
			labelValues[i] = curry[iCurry].value
			iCurry++
			continue
		}
		labelValues[i] = lvs[iLVs]
		iLVs++
	}
	return labelValues
}

func validateValuesInLabels(labels prometheus.Labels, expectedNumberOfValues int) error {
	if len(labels) != expectedNumberOfValues {
		return fmt.Errorf(
			"%s: expected %d label values but got %d in %#v",
			errors.New("inconsistent label cardinality"), expectedNumberOfValues,
			len(labels), labels,
		)
	}

	for name, val := range labels {
		if !utf8.ValidString(val) {
			return fmt.Errorf("label %s: value %q is not valid UTF-8", name, val)
		}
	}

	return nil
}

func validateLabelValues(vals []string, expectedNumberOfValues int) error {
	if len(vals) != expectedNumberOfValues {
		return fmt.Errorf(
			"%s: expected %d label values but got %d in %#v",
			errInconsistentCardinality, expectedNumberOfValues,
			len(vals), vals,
		)
	}

	for _, val := range vals {
		if !utf8.ValidString(val) {
			return fmt.Errorf("label value %q is not valid UTF-8", val)
		}
	}

	return nil
}

var errInconsistentCardinality = errors.New("inconsistent label cardinality")

func makeInconsistentCardinalityError(fqName string, labels, labelValues []string) error {
	return fmt.Errorf(
		"%s: %q has %d variable labels named %q but %d values %q were provided",
		errInconsistentCardinality, fqName,
		len(labels), labels,
		len(labelValues), labelValues,
	)
}
