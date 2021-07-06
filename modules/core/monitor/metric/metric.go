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

package metric

import "encoding/json"

// Metric .
type Metric struct {
	Name      string                 `json:"name"`
	Timestamp int64                  `json:"timestamp"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
}

func (m *Metric) String() string {
	bytes, _ := json.Marshal(m)
	return string(bytes)
}

// Copy instance
func (m *Metric) Copy() *Metric {
	tags := make(map[string]string)
	for k, v := range m.Tags {
		tags[k] = v
	}
	fields := make(map[string]interface{})
	for k, v := range m.Fields {
		fields[k] = v
	}
	copied := &Metric{
		Name:      m.Name,
		Timestamp: m.Timestamp,
		Tags:      tags,
		Fields:    fields,
	}
	return copied
}

// New .
func New() *Metric {
	return &Metric{
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}
}
