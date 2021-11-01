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

package persist

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/erda-project/erda/modules/core/monitor/metric"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
)

// Statistics .
type Statistics interface {
	storekit.ConsumeStatistics

	DecodeError(value []byte, err error)
	ValidateError(data *metric.Metric)
	MetadataError(data *metric.Metric, err error)
	MetadataUpdates(v int)
}

type statistics struct {
	readErrors    prometheus.Counter
	readBytes     *prometheus.CounterVec
	writeErrors   *prometheus.CounterVec
	confirmErrors *prometheus.CounterVec
	success       *prometheus.CounterVec

	decodeErrors   prometheus.Counter
	validateErrors *prometheus.CounterVec
	metadataError  *prometheus.CounterVec

	metadataUpdates prometheus.Counter
}

var sharedStatistics = newStatistics()

func newStatistics() Statistics {
	const subSystem = "metric_persist"
	s := &statistics{
		readErrors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name:      "read_errors",
				Subsystem: subSystem,
			},
		),
		readBytes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "read_bytes",
				Subsystem: subSystem,
			}, distinguishingKeys,
		),
		writeErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "write_errors",
				Subsystem: subSystem,
			}, distinguishingKeys,
		),
		confirmErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "confirm_errors",
				Subsystem: subSystem,
			}, distinguishingKeys,
		),
		success: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "success",
				Subsystem: subSystem,
			}, distinguishingKeys,
		),
		decodeErrors: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name:      "decode_errors",
				Subsystem: subSystem,
			},
		),
		validateErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "validate_errors",
				Subsystem: subSystem,
			}, distinguishingKeys,
		),
		metadataError: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name:      "metadata_errors",
				Subsystem: subSystem,
			}, distinguishingKeys,
		),
		metadataUpdates: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name:      "metadata_updates",
				Subsystem: subSystem,
			},
		),
	}

	// only register once
	prometheus.MustRegister(
		s.readErrors,
		s.readBytes,
		s.writeErrors,
		s.confirmErrors,
		s.success,
		s.decodeErrors,
		s.validateErrors,
		s.metadataError,
		s.metadataUpdates,
	)
	return s
}

func (s *statistics) ReadError(err error) {
	s.readErrors.Inc()
}

func (s *statistics) DecodeError(value []byte, err error) {
	s.decodeErrors.Inc()
}

func (s *statistics) WriteError(list []interface{}, err error) {
	for _, item := range list {
		s.writeErrors.WithLabelValues(getStatisticsLabels(item.(*metric.Metric))...).Inc()
	}
}

func (s *statistics) ConfirmError(list []interface{}, err error) {
	for _, item := range list {
		s.confirmErrors.WithLabelValues(getStatisticsLabels(item.(*metric.Metric))...).Inc()
	}
}

func (s *statistics) Success(list []interface{}) {
	for _, item := range list {
		s.success.WithLabelValues(getStatisticsLabels(item.(*metric.Metric))...).Inc()
	}
}

func (s *statistics) ValidateError(data *metric.Metric) {
	s.validateErrors.WithLabelValues(getStatisticsLabels(data)...).Inc()
}

func (s *statistics) MetadataError(data *metric.Metric, err error) {
	s.metadataError.WithLabelValues(getStatisticsLabels(data)...).Inc()
}

func (s *statistics) MetadataUpdates(v int) {
	s.metadataUpdates.Add(float64(v))
}

var distinguishingKeys = []string{
	"metric",
	"org_name", "cluster_name",
}

func getStatisticsLabels(data *metric.Metric) []string {
	return []string{
		data.Name,
		data.Tags["org_name"],
		data.Tags["cluster_name"],
	}
}
