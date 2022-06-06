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
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/erda-project/erda/modules/apps/msp/apm/exception/model"
	"github.com/erda-project/erda/modules/tools/monitor/core/storekit"
)

// Statistics .
type Statistics interface {
	storekit.ConsumeStatistics

	DecodeError(value []byte, err error)
	ValidateError(data *model.Event)
	MetadataError(data *model.Event, err error)
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

	// performance
	readLatency  prometheus.Histogram
	writeLatency prometheus.Histogram
}

var sharedStatistics = newStatistics()

func newStatistics() Statistics {
	const subSystem = "error_event_persist"
	s := &statistics{
		readLatency: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:      "read_latency",
				Subsystem: subSystem,
				Buckets:   storekit.DefaultLatencyBuckets,
			},
		),
		writeLatency: prometheus.NewHistogram(
			prometheus.HistogramOpts{
				Name:      "write_latency",
				Subsystem: subSystem,
				Buckets:   storekit.DefaultLatencyBuckets,
			},
		),
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
	}

	// only register once
	prometheus.MustRegister(
		s.readLatency,
		s.writeLatency,
		s.readErrors,
		s.readBytes,
		s.writeErrors,
		s.confirmErrors,
		s.success,
		s.decodeErrors,
		s.validateErrors,
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
		s.writeErrors.WithLabelValues(getStatisticsLabels(item.(*model.Event))...).Inc()
	}
}

func (s *statistics) ConfirmError(list []interface{}, err error) {
	for _, item := range list {
		s.confirmErrors.WithLabelValues(getStatisticsLabels(item.(*model.Event))...).Inc()
	}
}

func (s *statistics) Success(list []interface{}) {
	for _, item := range list {
		s.success.WithLabelValues(getStatisticsLabels(item.(*model.Event))...).Inc()
	}
}

func (s *statistics) ValidateError(data *model.Event) {
	s.validateErrors.WithLabelValues(getStatisticsLabels(data)...).Inc()
}

func (*statistics) MetadataError(data *model.Event, err error) {}

func (s *statistics) ObserveReadLatency(start time.Time) {
	s.readLatency.Observe(float64(time.Since(start).Milliseconds()))
}

func (s *statistics) ObserveWriteLatency(start time.Time) {
	s.writeLatency.Observe(float64(time.Since(start).Milliseconds()))
}

var distinguishingKeys = []string{
	"org_name", "cluster_name",
	// "scope", "scope_id",
}

func getStatisticsLabels(data *model.Event) []string {
	// var scope, scopeID string
	//
	// if app, ok := data.Tags["application_name"]; ok {
	// 	scope = "app"
	// 	if project, ok := data.Tags["project_name"]; ok {
	// 		scopeID = project + "/" + app
	// 	} else {
	// 		scopeID = app
	// 	}
	// }
	return []string{
		data.Tags["org_name"],
		data.Tags["cluster_name"],
		// scope, scopeID,
	}
}
