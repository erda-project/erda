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

package storage

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
)

// Metric .
type Metric struct {
	*metrics.Metric
	Date int64 `json:"@timestamp"`
}

func (p *provider) invoke(key []byte, value []byte, topic *string, timestamp time.Time) error {
	m := &Metric{}
	if err := json.Unmarshal(value, m); err != nil {
		return err
	}

	if len(p.C.Output.Features.FilterPrefix) > 0 { // 暂时简单过滤一下
		if strings.HasPrefix(m.Name, p.C.Output.Features.FilterPrefix) {
			return nil
		}
	}

	if m.Tags == nil || m.Tags[MetricTagLifetime] == MetricTagTransient {
		return nil
	}
	if p.C.Output.Features.GenerateMeta {
		if err := p.metaProcessor.add(m); err != nil {
			// 转换 meta 失败，不堵塞存储原始 metric
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
			// 如果索引不存在，推送到另外的 topic 中去处理
			p.output.kafka.Write(&kafka.Message{
				Topic: &p.C.Output.Kafka.Topic,
				Data:  bytes,
				Key:   []byte(m.Name), // 相同key会路由到相同的partition，尽量保证一种索引由一个结点创建
			})
			return nil
		}
		documentIndex = index
	}
	return p.output.es.Write(&elasticsearch.Document{Index: documentIndex, ID: documentID, Data: m})
}

// 处理索引创建，并且写入es，该处理函数比invoke的指标处理速度要慢，和invoke的指标处理分开，尽量避免索引创建等操作影响其他指标数据写入。
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
	MetricTagLifetime    = "_lt"       //lifetime
	MetricTagTransient   = "transient" //瞬态，不存储到es

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
