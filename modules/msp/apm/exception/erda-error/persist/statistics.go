package persist

import (
	"github.com/erda-project/erda/modules/core/monitor/storekit"
	"github.com/erda-project/erda/modules/msp/apm/exception"
	"github.com/prometheus/client_golang/prometheus"
)

// Statistics .
type Statistics interface {
	storekit.ConsumeStatistics

	DecodeError(value []byte, err error)
	ValidateError(data *exception.Erda_error)
	MetadataError(data *exception.Erda_error, err error)
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
}

var sharedStatistics = newStatistics()

func newStatistics() Statistics {
	const subSystem = "error_persist"
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
		s.writeErrors.WithLabelValues(getStatisticsLabels(item.(*exception.Erda_error))...).Inc()
	}
}

func (s *statistics) ConfirmError(list []interface{}, err error) {
	for _, item := range list {
		s.confirmErrors.WithLabelValues(getStatisticsLabels(item.(*exception.Erda_error))...).Inc()
	}
}

func (s *statistics) Success(list []interface{}) {
	for _, item := range list {
		s.success.WithLabelValues(getStatisticsLabels(item.(*exception.Erda_error))...).Inc()
	}
}

func (s *statistics) ValidateError(data *exception.Erda_error) {
	s.validateErrors.WithLabelValues(getStatisticsLabels(data)...).Inc()
}

func (*statistics) MetadataError(data *exception.Erda_error, err error) {}

var distinguishingKeys = []string{
	"org_name", "cluster_name",
	"scope", "scope_id",
}

func getStatisticsLabels(data *exception.Erda_error) []string {
	var scope, scopeID string

	if app, ok := data.Tags["application_name"]; ok {
		scope = "app"
		if project, ok := data.Tags["project_name"]; ok {
			scopeID = project + "/" + app
		} else {
			scopeID = app
		}
	}
	return []string{
		data.Tags["org_name"],
		data.Tags["cluster_name"],
		scope, scopeID,
	}
}
