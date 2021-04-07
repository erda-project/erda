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

package metrics

import (
	"encoding/json"
	"sort"
)

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
	newMetric := &Metric{
		Name:      m.Name,
		Timestamp: m.Timestamp,
		Tags:      tags,
		Fields:    fields,
	}
	return newMetric
}

// New .
func New() *Metric {
	return &Metric{
		Tags:   make(map[string]string),
		Fields: make(map[string]interface{}),
	}
}

// FieldDefine .
type FieldDefine struct {
	Key    string         `json:"key"`
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Unit   string         `json:"uint"`
	Values []*ValueDefine `json:"values,omitempty"`
}

// TagDefine .
type TagDefine struct {
	Key    string         `json:"key"`
	Name   string         `json:"name"`
	Values []*ValueDefine `json:"values,omitempty"`
}

// ValueDefine .
type ValueDefine struct {
	Value interface{} `json:"value"`
	Name  string      `json:"name"`
}

// NameDefine .
type NameDefine struct {
	Key  string `json:"key" `
	Name string `json:"name"`
}

// MetricMeta .
type MetricMeta struct {
	Name   NameDefine              `json:"name"`
	Labels map[string]string       `json:"labels,omitempty"`
	Tags   map[string]*TagDefine   `json:"tags"`
	Fields map[string]*FieldDefine `json:"fields"`
}

// NewMeta .
func NewMeta() *MetricMeta {
	return &MetricMeta{
		Tags:   make(map[string]*TagDefine),
		Fields: make(map[string]*FieldDefine),
	}
}

// TagsKeys .
func (m *MetricMeta) TagsKeys() []string {
	var keys []string
	for k := range m.Tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// FieldsKeys .
func (m *MetricMeta) FieldsKeys() []string {
	var keys []string
	for k := range m.Fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
