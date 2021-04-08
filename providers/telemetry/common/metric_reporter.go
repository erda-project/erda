// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package common

import (
	"encoding/json"
	"net"
	"os"
)

var (
	DEFAULT_REPORTER, _ = NewMetricReporter()
	DEFAULT_BUCKET      = 10
)

type MetricReporter struct {
	conn   net.Conn
	labels GlobalLabel
}

func NewMetricReporter() (*MetricReporter, error) {
	hostIp := os.Getenv("HOST_IP")
	if hostIp == "" {
		hostIp = "localhost"
	}
	hostPort := os.Getenv("HOST_PORT")
	if hostPort == "" {
		hostPort = "7082"
	}
	conn, err := net.Dial("udp", hostIp+":"+hostPort)
	if err != nil {
		return nil, err
	}
	return &MetricReporter{
		conn:   conn,
		labels: GetGlobalLabels(),
	}, nil
}

func (m *MetricReporter) Report(metrics []*Metric) (err error) {
	length := len(metrics)

	if length == 0 {
		return
	}

	if length <= DEFAULT_BUCKET {
		return m.send(metrics)
	}

	idx := 0
	for {
		bucket := DEFAULT_BUCKET
		if length-idx < DEFAULT_BUCKET {
			bucket = length - idx
		}
		end := idx + bucket
		bucketMetrics := metrics[idx:end]
		err = m.send(bucketMetrics)
		idx = end
		if idx >= length {
			break
		}
	}
	return err
}

func (m *MetricReporter) send(metrics []*Metric) (err error) {
	for _, metric := range metrics {
		for k, v := range m.labels {
			if _, ok := metric.Tags[k]; !ok {
				metric.Tags[k] = v
			}
		}
	}

	if data, err := json.Marshal(metrics); err == nil {
		if m.conn != nil {
			_, err = m.conn.Write(data)
		}
	}
	return err
}
