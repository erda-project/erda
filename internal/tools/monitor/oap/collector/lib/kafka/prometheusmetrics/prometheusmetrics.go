// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prometheusmetrics

// This code is based on a code of https://github.com/deathowl/go-metrics-prometheus library.

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rcrowley/go-metrics"
)

type exporter struct {
	opt              Options
	registry         MetricsRegistry
	promRegistry     prometheus.Registerer
	gauges           map[string]prometheus.Gauge
	customMetrics    map[string]*customCollector
	histogramBuckets []float64
	timerBuckets     []float64
	mutex            *sync.Mutex
}

func (c *exporter) sanitizeName(key string) string {
	ret := []byte(key)
	for i := 0; i < len(ret); i++ {
		c := key[i]
		allowed := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == ':' || (c >= '0' && c <= '9')
		if !allowed {
			ret[i] = '_'
		}
	}
	return string(ret)
}

func (c *exporter) createKey(name string) string {
	return c.opt.Namespace + "_" + c.opt.Subsystem + "_" + name
}

func (c *exporter) gaugeFromNameAndValue(name string, val float64) error {
	shortName, labels, skip := c.metricNameAndLabels(name)
	if skip {
		if c.opt.Debug {
			fmt.Printf("[saramaprom] skip metric %q because there is no broker or topic labels\n", name)
		}
		return nil
	}

	if _, exists := c.gauges[name]; !exists {
		labelNames := make([]string, 0, len(labels))
		for labelName := range labels {
			labelNames = append(labelNames, labelName)
		}

		g := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace:   c.sanitizeName(c.opt.Namespace),
			Subsystem:   c.sanitizeName(c.opt.Subsystem),
			Name:        c.sanitizeName(shortName),
			Help:        shortName,
			ConstLabels: c.opt.ConstantLabels,
		}, labelNames)

		if err := c.promRegistry.Register(g); err != nil {
			switch err := err.(type) {
			case prometheus.AlreadyRegisteredError:
				var ok bool
				g, ok = err.ExistingCollector.(*prometheus.GaugeVec)
				if !ok {
					return fmt.Errorf("prometheus collector already registered but it's not *prometheus.GaugeVec: %v", g)
				}
			default:
				return err
			}
		}
		c.gauges[name] = g.With(labels)
	}

	c.gauges[name].Set(val)
	return nil
}

func (c *exporter) metricNameAndLabels(metricName string) (newName string, labels map[string]string, skip bool) {
	newName, broker, topic := parseMetricName(metricName)
	if broker == "" && topic == "" {
		// skip metrics for total
		return newName, labels, true
	}
	labels = map[string]string{}
	if broker != "" {
		labels["broker"] = broker
	}
	if topic != "" {
		labels["topic"] = broker
	}
	if c.opt.Label != "" {
		labels["label"] = c.opt.Label
	}
	return newName, labels, false
}

func parseMetricName(name string) (newName, broker, topic string) {
	if i := strings.Index(name, "-for-broker-"); i >= 0 {
		newName = name[:i]
		broker = name[i+len("-for-broker-"):]
		return
	}
	if i := strings.Index(name, "-for-topic-"); i >= 0 {
		newName = name[:i]
		topic = name[i+len("-for-topic-"):]
		return
	}
	return name, "", ""
}

func (c *exporter) histogramFromNameAndMetric(name string, goMetric interface{}, buckets []float64) error {
	key := c.createKey(name)
	collector, exists := c.customMetrics[key]
	if !exists {
		collector = newCustomCollector(c.mutex)
		c.promRegistry.MustRegister(collector)
		c.customMetrics[key] = collector
	}

	var ps []float64
	var count uint64
	var sum float64
	var typeName string

	switch metric := goMetric.(type) {
	case metrics.Histogram:
		snapshot := metric.Snapshot()
		ps = snapshot.Percentiles(buckets)
		count = uint64(snapshot.Count())
		sum = float64(snapshot.Sum())
		typeName = "histogram"
	case metrics.Timer:
		snapshot := metric.Snapshot()
		ps = snapshot.Percentiles(buckets)
		count = uint64(snapshot.Count())
		sum = float64(snapshot.Sum())
		typeName = "timer"
	default:
		return fmt.Errorf("unexpected metric type %T", goMetric)
	}

	bucketVals := make(map[float64]uint64)
	for ii, bucket := range buckets {
		bucketVals[bucket] = uint64(ps[ii])
	}

	name, labels, skip := c.metricNameAndLabels(name)
	if skip {
		return nil
	}

	for k, v := range c.opt.ConstantLabels {
		labels[k] = v
	}
	desc := prometheus.NewDesc(
		prometheus.BuildFQName(
			c.sanitizeName(c.opt.Namespace),
			c.sanitizeName(c.opt.Subsystem),
			c.sanitizeName(name)+"_"+typeName,
		),
		c.sanitizeName(name),
		nil,
		labels,
	)

	hist, err := prometheus.NewConstHistogram(desc, count, sum, bucketVals)
	if err != nil {
		return err
	}
	c.mutex.Lock()
	collector.metric = hist
	c.mutex.Unlock()
	return nil
}

func (c *exporter) update() error {
	if c.opt.Debug {
		fmt.Print("[saramaprom] update()\n")
	}
	var err error
	c.registry.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Counter:
			err = c.gaugeFromNameAndValue(name, float64(metric.Count()))
		case metrics.Gauge:
			err = c.gaugeFromNameAndValue(name, float64(metric.Value()))
		case metrics.GaugeFloat64:
			err = c.gaugeFromNameAndValue(name, float64(metric.Value()))
		case metrics.Histogram: // sarama
			samples := metric.Snapshot().Sample().Values()
			if len(samples) > 0 {
				lastSample := samples[len(samples)-1]
				err = c.gaugeFromNameAndValue(name, float64(lastSample))
			}
			if err == nil {
				err = c.histogramFromNameAndMetric(name, metric, c.histogramBuckets)
			}
		case metrics.Meter: // sarama
			lastSample := metric.Snapshot().Rate1()
			err = c.gaugeFromNameAndValue(name, float64(lastSample))
		case metrics.Timer:
			lastSample := metric.Snapshot().Rate1()
			err = c.gaugeFromNameAndValue(name, float64(lastSample))
			if err == nil {
				err = c.histogramFromNameAndMetric(name, metric, c.timerBuckets)
			}
		}
	})
	return err
}

// for collecting prometheus.constHistogram objects
type customCollector struct {
	prometheus.Collector

	metric prometheus.Metric
	mutex  *sync.Mutex
}

func newCustomCollector(mu *sync.Mutex) *customCollector {
	return &customCollector{
		mutex: mu,
	}
}

func (c *customCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	if c.metric != nil {
		val := c.metric
		ch <- val
	}
	c.mutex.Unlock()
}

func (c *customCollector) Describe(_ chan<- *prometheus.Desc) {
	// empty method to fulfill prometheus.Collector interface
}

// Options holds optional params for ExportMetrics.
type Options struct {
	// PrometheusRegistry is prometheus registry. Default prometheus.DefaultRegisterer.
	PrometheusRegistry prometheus.Registerer

	// Namespace and Subsystem form the metric name prefix.
	// Default Subsystem is "sarama".
	Namespace string
	Subsystem string

	// Label specifies value of "label" label. Default "".
	Label string

	ConstantLabels map[string]string

	// FlushInterval specifies interval between updating metrics. Default 1s.
	FlushInterval time.Duration

	// OnError is error handler. Default handler panics when error occurred.
	OnError func(err error)

	// Debug turns on debug logging.
	Debug bool
}

// ExportMetrics exports metrics from go-metrics to prometheus.
func ExportMetrics(ctx context.Context, metricsRegistry MetricsRegistry, opt Options) error {
	if opt.PrometheusRegistry == nil {
		opt.PrometheusRegistry = prometheus.DefaultRegisterer
	}
	if opt.Subsystem == "" {
		opt.Subsystem = "sarama"
	}
	if opt.FlushInterval == 0 {
		opt.FlushInterval = 10 * time.Second
	}
	if opt.OnError != nil {
		opt.OnError = func(err error) {
			panic(fmt.Errorf("saramaprom: %w", err))
		}
	}

	exp := &exporter{
		opt:              opt,
		registry:         metricsRegistry,
		promRegistry:     opt.PrometheusRegistry,
		gauges:           make(map[string]prometheus.Gauge),
		customMetrics:    make(map[string]*customCollector),
		histogramBuckets: []float64{0.05, 0.1, 0.25, 0.50, 0.75, 0.9, 0.95, 0.99},
		timerBuckets:     []float64{0.50, 0.95, 0.99, 0.999},
		mutex:            new(sync.Mutex),
	}

	err := exp.update()
	if err != nil {
		return fmt.Errorf("saramaprom: %w", err)
	}

	go func() {
		t := time.NewTicker(opt.FlushInterval)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				err := exp.update()
				if err != nil {
					opt.OnError(err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

// MetricsRegistry is an interface for 'github.com/rcrowley/go-metrics'.Registry
// which is used for metrics in sarama.
type MetricsRegistry interface {
	Each(func(name string, i interface{}))
}
