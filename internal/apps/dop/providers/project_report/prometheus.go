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

package project_report

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
)

type metricValue struct {
	value     float64
	labels    []string
	timestamp time.Time
}

type metricValues []metricValue

type iterationMetric struct {
	name              string
	help              string
	valueType         prometheus.ValueType
	extraLabels       []string
	condition         func(i *apistructs.Iteration) bool
	getValues         func(i *IterationInfo) metricValues
	getMetricsItemIDs func(i *IterationInfo) string
}

func (im *iterationMetric) desc(baseLabels []string) *prometheus.Desc {
	return prometheus.NewDesc(im.name, im.help, append(baseLabels, im.extraLabels...), nil)
}

type IterationLabelsFunc func(info *IterationInfo) map[string]string

type infoProvider interface {
	// GetRequestedIterationsInfo gets info for all requested iterations based on the request options.
	GetRequestedIterationsInfo() (map[uint64]*IterationInfo, error)
}

type PrometheusCollector struct {
	infoProvider        infoProvider
	errors              prometheus.Gauge
	iterationMetrics    []iterationMetric
	iterationLabelsFunc IterationLabelsFunc
}

func (c *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	c.errors.Set(0)
	c.collectIterationInfo(ch)
	c.errors.Collect(ch)
}

func (i *iterationCollector) Collect(ch chan<- prometheus.Metric) {
	i.helper.errors.Set(0)
	iterations, err := i.helper.infoProvider.GetRequestedIterationsInfo()
	if err != nil {
		i.helper.errors.Set(1)
		logrus.Errorf("failed to get iteration info: %v", err)
	}

	for _, iter := range iterations {
		if iter.IterationMetricFields == nil {
			continue
		}
		rawLabels := map[string]struct{}{}
		for l := range i.iterationIDsLabelsFunc(iter) {
			rawLabels[l] = struct{}{}
		}
		values := make([]string, 0, len(rawLabels))
		labels := make([]string, 0, len(rawLabels))
		iterationLabels := i.iterationIDsLabelsFunc(iter)
		for l := range rawLabels {
			duplicate := false
			sl := sanitizeLabelName(l)
			for _, x := range labels {
				if sl == x {
					duplicate = true
					break
				}
			}
			if !duplicate {
				labels = append(labels, sl)
				values = append(values, iterationLabels[sl])
			}
		}

		for _, im := range i.helper.iterationMetrics {
			for k, v := range labels {
				if v == "metrics_type" {
					values[k] = im.name
				}
				if v == "ids" {
					values[k] = im.getMetricsItemIDs(iter)
				}
			}
			desc := im.desc(labels)
			for _, metricVal := range im.getValues(iter) {
				ch <- prometheus.NewMetricWithTimestamp(
					metricVal.timestamp,
					prometheus.MustNewConstMetric(desc, im.valueType, metricVal.value, append(values, metricVal.labels...)...),
				)
			}
		}
	}
	i.helper.errors.Collect(ch)
}

func (c *PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	c.errors.Describe(ch)
	for _, im := range c.iterationMetrics {
		ch <- im.desc([]string{})
	}
}

func (i *iterationCollector) Describe(ch chan<- *prometheus.Desc) {
	i.helper.errors.Describe(ch)
	for _, im := range i.helper.iterationMetrics {
		ch <- im.desc([]string{})
	}
}

func (c *PrometheusCollector) collectIterationInfo(ch chan<- prometheus.Metric) {
	iterations, err := c.infoProvider.GetRequestedIterationsInfo()
	if err != nil {
		c.errors.Set(1)
		logrus.Errorf("failed to get iteration info: %v", err)
	}

	for _, iter := range iterations {
		iterationLabels := c.iterationLabelsFunc(iter)
		if iter.IterationMetricFields != nil {
			iterationLabels[labelIterationItemUUID] = iter.IterationMetricFields.UUID
		}
		rawLabels := map[string]struct{}{}
		for l := range iterationLabels {
			rawLabels[l] = struct{}{}
		}
		values := make([]string, 0, len(rawLabels))
		labels := make([]string, 0, len(rawLabels))

		for l := range rawLabels {
			duplicate := false
			sl := sanitizeLabelName(l)
			for _, x := range labels {
				if sl == x {
					duplicate = true
					break
				}
			}
			if !duplicate {
				labels = append(labels, sl)
				values = append(values, iterationLabels[sl])
			}
		}

		for _, im := range c.iterationMetrics {
			desc := im.desc(labels)
			for _, metricVal := range im.getValues(iter) {
				ch <- prometheus.NewMetricWithTimestamp(
					metricVal.timestamp,
					prometheus.MustNewConstMetric(desc, im.valueType, metricVal.value, append(values, metricVal.labels...)...),
				)
			}
		}
	}
}
