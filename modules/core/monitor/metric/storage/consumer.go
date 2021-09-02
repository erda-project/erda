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

package storage

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/modules/core/monitor/metric"
)

// Metric .
type Metric struct {
	*metric.Metric
	Date int64 `json:"@timestamp"`
}

func (p *provider) invoke(key []byte, value []byte, topic *string, timestamp time.Time) error {
	m := &Metric{}
	if err := json.Unmarshal(value, m); err != nil {
		return err
	}

	if len(p.C.Output.Features.FilterPrefix) > 0 { // Just a little filter for now.
		if strings.HasPrefix(m.Name, p.C.Output.Features.FilterPrefix) {
			return nil
		}
	}

	if m.Tags == nil || m.Tags[MetricTagLifetime] == MetricTagTransient {
		return nil
	}
	if p.C.Output.Features.GenerateMeta {
		if err := p.metaProcessor.add(m); err != nil {
			// Convert Meta failed, do not block stored original metric.
			p.L.Errorf("store metric[%s] meta error", m.Name)
		}
	}
	processInvalidFields(m)
	processTimestampDateFormat(m)
	// if p.C.Output.Features.Counter {
	// 	countMetric(p.counter, m.Metric)
	// }
	if p.C.Output.Features.MachineSummary {
		if doc, ok := p.handleMachineSummary(m); ok {
			return p.output.es.Write(doc)
		}
	}
	var documentIndex, documentID string
	if ttl, ok := m.Tags[MetricTagTTL]; ok && ttl == MetricTagTTLFixed {
		documentIndex = p.index.GetWriteFixedIndex(m.Metric)
		if id, ok := m.Tags[MetricTagMetricID]; ok {
			documentID = id
		}
	} else if id, ok := m.Tags[MetricTagMetricID]; ok {
		documentIndex = p.index.GetWriteFixedIndex(m.Metric)
		documentID = id
	} else {
		index, ok := p.index.GetWriteIndex(m.Metric)
		if !ok {
			bytes, _ := json.Marshal(m)
			// If the index does not exist, push it to another topic for processing.
			p.output.kafka.Write(&kafka.Message{
				Topic: &p.C.Output.Kafka.Topic,
				Data:  bytes,
				Key:   []byte(m.Name), // The same key will be routed to the same partition, so as to ensure that one index is created by one node.
			})
			return nil
		}
		documentIndex = index
	}
	return p.output.es.Write(&elasticsearch.Document{Index: documentIndex, ID: documentID, Data: m})
}

// The same key will be routed to the same partition, so as to ensure that one index is created by one node.
func (p *provider) handleCreatingIndexMetric(key []byte, value []byte, topic *string, timestamp time.Time) error {
	m := &Metric{}
	if err := json.Unmarshal(value, m); err != nil {
		return err
	}
	if m.Tags == nil || m.Tags[MetricTagLifetime] == MetricTagTransient {
		return nil
	}
	err := p.index.CreateIndex(m.Metric)
	if err != nil {
		return err
	}
	index, _ := p.index.GetWriteIndex(m.Metric)
	return p.output.es.Write(&elasticsearch.Document{Index: index, Data: m})
}

//
const (
	MetricInternalPrefix = "_"
	MetricMeta           = "_metric_meta"

	MetricTagMetricName  = "metric_name"
	MetricTagScope       = "metric_scope"
	MetricTagScopeID     = "metric_scope_id"
	MetricTagMetricMeta  = "meta"
	MetricTagMetricID    = "_id"
	MetricTagOrgName     = "org_name"
	MetricTagClusterName = "cluster_name"
	MetricTagTTL         = "_ttl"
	MetricTagTTLFixed    = "fixed"
	MetricTagLifetime    = "_lt"       // lifetime
	MetricTagTransient   = "transient" // transient, not stored to elasticsearch.

	MetricFieldTags   = "tags"
	MetricFieldFields = "fields"

	Minute int64 = Second * 60
	Second int64 = 1000 * 1000 * 1000

	srcPrefix         = "src_"
	srcClusterNameKey = srcPrefix + MetricTagClusterName
	srcOrgNameKey     = srcPrefix + MetricTagOrgName
	metricNameKey     = MetricTagMetricName
	metricScopeKey    = MetricTagScope
	metricScopeIDKey  = MetricTagScopeID
)

const esMaxValue = float64(math.MaxInt64)

func processInvalidFields(m *Metric) {
	fields := m.Fields
	if fields == nil {
		return
	}
	for k, v := range fields {
		switch val := v.(type) {
		case float64:
			if val > esMaxValue {
				fields[k] = strconv.FormatFloat(val, 'f', -1, 64)
			}
		}
	}
}

func processTimestampDateFormat(m *Metric) *Metric {
	if len(strconv.FormatInt(m.Timestamp, 10)) > 13 {
		m.Date = m.Timestamp / int64(time.Millisecond)
	} else {
		m.Date = m.Timestamp
	}
	return m
}

func (p *provider) handleMachineSummary(metric *Metric) (*elasticsearch.Document, bool) {
	if metric.Name != "machine_summary" {
		return nil, false
	}
	if labels, ok := metric.Tags["labels"]; ok {
		metric.Fields["labels"] = strings.Split(labels, ",")
	}
	var (
		id    string
		hasID bool
	)
	if id, hasID = metric.Tags["terminus_index_id"]; !hasID {
		id = metric.Tags["cluster_name"] + "/" + metric.Tags["host_ip"]
	}
	return &elasticsearch.Document{Index: p.index.GetWriteFixedIndex(metric.Metric), ID: id, Data: metric}, true
}

// func countMetric(counter *promxp.AutoResetCounterVec, metric *metrics.Metric) {
// 	counter.WithLabelValues(
// 		metric.Name,
// 		// metric.Tags["_"+MetricTagScope],
// 		// metric.Tags["_"+MetricTagScopeID],
// 		metric.Tags[MetricTagOrgName],
// 		metric.Tags[MetricTagClusterName]).
// 		Inc()
// }
