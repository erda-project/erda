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

import (
	"encoding/json"
	"hash/fnv"
	"sort"

	"github.com/erda-project/erda/pkg/strutil"
)

const (
	// distributed table
	CH_TABLE_ALL = "metrics_all"
	CH_TABLE     = "metrics"
)

// Metric .
type Metric struct {
	Name      string                 `json:"name"`
	Timestamp int64                  `json:"timestamp"`
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
	OrgName   string                 `json:"-"`
}

func (m *Metric) Hash() uint64 {
	sortedTags := make([]tag, len(m.Tags))
	for k, v := range m.Tags {
		sortedTags = append(sortedTags, tag{k, v})
	}
	sort.Slice(sortedTags, func(i, j int) bool {
		return sortedTags[i].key < sortedTags[j].key
	})

	h := fnv.New64a()
	for _, item := range sortedTags {
		h.Write(strutil.NoCopyStringToBytes(item.key))
		h.Write(strutil.NoCopyStringToBytes("\n"))
		h.Write(strutil.NoCopyStringToBytes(item.value))
		h.Write(strutil.NoCopyStringToBytes("\n"))
	}
	h.Write(strutil.NoCopyStringToBytes(m.Name))
	return h.Sum64()
}

type tag struct {
	key, value string
}

func (m *Metric) GetTags() map[string]string {
	if m.Tags == nil {
		m.Tags = map[string]string{}
	}
	return m.Tags
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

type TableMetrics struct {
	OrgName           string    `ch:"org_name"`
	MetricGroup       string    `ch:"metric_group"`
	Timestamp         int64     `ch:"timestamp"`
	NumberFieldKeys   []string  `ch:"number_field_keys"`
	NumberFieldValues []float64 `ch:"number_field_values"`
	StringFieldKeys   []string  `ch:"string_field_keys"`
	StringFieldValues []string  `ch:"string_field_values"`
	TagKeys           []string  `ch:"tag_keys"`
	TagValues         []string  `ch:"tag_values"`
}
