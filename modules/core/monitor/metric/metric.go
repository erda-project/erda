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
