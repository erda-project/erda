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
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda/modules/core/monitor/metric"
	indexmanager "github.com/erda-project/erda/modules/core/monitor/metric/index"
	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/erda-project/erda/pkg/arrays"
)

type metricMetaCache struct {
	lastTimestamp int64
	metric        *metric.Metric
}

type metaProcessor struct {
	metrics   chan *metric.Metric
	metaCache map[string]*metricMetaCache

	es      writer.Writer
	index   indexmanager.Index
	counter *prometheus.Counter

	r *rand.Rand
}

func createMetaProcess(es writer.Writer, index indexmanager.Index, counter *prometheus.Counter) *metaProcessor {
	p := &metaProcessor{
		metrics:   make(chan *metric.Metric, 64),
		metaCache: make(map[string]*metricMetaCache),
		es:        es,
		index:     index,
		counter:   counter,
		r:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	go p.process()
	return p
}

func (p *metaProcessor) add(m *Metric) error {

	// tags 包含 _meta 并且值为 "true"，转换为 meta 发送到 kafka
	// {
	// 	"name":"xxx",
	// 	"tags":{
	// 		"_meta": "true"
	// 	}
	// }

	// 转换后的 metric_meta
	// {
	// 	"name":"_metric_meta",
	// 	"tags":{
	// 		"metric_name": "application_http"
	// 	},
	// 	"fields": {
	// 		"tags": [
	// 			"service_name",
	// 			"terminus_key"
	// 		],
	// 		"fields":[
	// 			"count:number"
	// 		]
	// 	}
	// }

	if m.Name == MetricMeta {
		return nil
	}

	if metaVal, ok := m.Tags[MetricInternalPrefix+MetricTagMetricMeta]; ok {
		if b, err := strconv.ParseBool(metaVal); err == nil && b {
			tags := make([]string, 0)
			fields := make([]string, 0)
			if m.Tags != nil && len(m.Tags) > 0 {
				for tag := range m.Tags {
					// Filter tags with built-in _ prefixes, such as _id _meta.
					if !strings.HasPrefix(tag, MetricInternalPrefix) {
						tags = append(tags, tag)
					}
				}
			}
			if m.Fields != nil && len(m.Fields) > 0 {
				for field, val := range m.Fields {
					fields = append(fields, field+":"+utils.TypeOf(val))
				}
			}

			metaId := m.Name + "-" +
				m.Tags["_"+MetricTagScope] + "-" + m.Tags["_"+MetricTagScopeID]

			metaTags := map[string]string{}
			if m.Tags != nil && len(m.Tags) > 0 {
				for tag, val := range m.Tags {
					// Filter tags with built-in _ prefixes, such as _id _meta.
					if strings.HasPrefix(tag, MetricInternalPrefix) {
						metaTags[tag[1:]] = val
					}
				}
			}
			metaTags[MetricTagMetricName] = m.Name
			metaTags[MetricTagMetricID] = metaId
			metaTags[MetricTagOrgName] = m.Tags[MetricTagOrgName]
			metaTags[MetricTagClusterName] = m.Tags[MetricTagClusterName]
			metaMetric := &metric.Metric{
				Name:      MetricMeta,
				Timestamp: m.Timestamp,
				Tags:      metaTags,
				Fields: map[string]interface{}{
					MetricFieldTags:   tags,
					MetricFieldFields: fields,
				},
			}
			p.metrics <- metaMetric
		}
	}

	return nil
}

func (p *metaProcessor) process() {
	for {
		metric, ok := <-p.metrics
		if ok {
			_ = p.processMetricMeta(metric)
		}
	}
}

func (p *metaProcessor) processMetricMeta(metric *metric.Metric) error {
	metaID := metric.Tags[MetricTagMetricID]
	metaCache, hasVal := p.metaCache[metaID]
	if !hasVal {
		metaCache = &metricMetaCache{
			metric: metric,
		}
		p.metaCache[metaID] = metaCache
	} else {
		tags := arrays.Distinct(arrays.Concat(metric.Fields[MetricFieldTags].([]string), metaCache.metric.Fields[MetricFieldTags].([]string)))
		fields := arrays.Distinct(arrays.Concat(metric.Fields[MetricFieldFields].([]string), metaCache.metric.Fields[MetricFieldFields].([]string)))
		metaCache.metric.Fields[MetricFieldTags] = tags
		metaCache.metric.Fields[MetricFieldFields] = fields
	}
	nowTimestamp := time.Now().UnixNano()
	if metaCache.lastTimestamp == 0 || nowTimestamp > metaCache.lastTimestamp {
		metaCache.lastTimestamp = nowTimestamp - nowTimestamp%Minute + Minute*15 - Second*int64(p.r.Intn(300))
		// p.L.Infof("push metric meta %s at %d", metric.Name, metaCache.lastTimestamp)
		// countMetric(p.counter, metaCache.metric)
		m := &Metric{}
		m.Metric = metaCache.metric.Copy()
		processTimestampDateFormat(m)
		return p.es.Write(&elasticsearch.Document{Index: p.index.GetWriteFixedIndex(metaCache.metric), ID: metaID, Data: m})
	}
	// return p.output.kafka.Write(meta)
	return nil
}
