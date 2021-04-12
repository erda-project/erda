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

package report

import "time"

type Metric struct {
	Name      string                 `json:"name"`
	Timestamp int64                  `json:"timestamp"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
}

type BulkMetricRequest []*Metric

func CreateBulkMetricRequest() *BulkMetricRequest {
	metricRequest := make(BulkMetricRequest, 0)
	return &metricRequest
}

func (b *BulkMetricRequest) Add(name string, tags map[string]string, fields map[string]interface{}) *BulkMetricRequest {
	*b = append(*b, &Metric{
		Name:      name,
		Timestamp: time.Now().UnixNano(),
		Tags:      tags,
		Fields:    fields,
	})
	return b
}

func (b *BulkMetricRequest) AddWithTime(name string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) *BulkMetricRequest {
	*b = append(*b, &Metric{
		Name:      name,
		Timestamp: timestamp.UnixNano(),
		Tags:      tags,
		Fields:    fields,
	})
	return b
}
