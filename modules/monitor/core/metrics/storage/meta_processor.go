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
	"math/rand"
	"strconv"
	"strings"
	"time"

	writer "github.com/erda-project/erda-infra/pkg/parallel-writer"
	"github.com/erda-project/erda-infra/providers/elasticsearch"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
	indexmanager "github.com/erda-project/erda/modules/monitor/core/metrics/index"
	"github.com/erda-project/erda/modules/monitor/utils"
	"github.com/erda-project/erda/pkg/arrays"
	"github.com/prometheus/client_golang/prometheus"
)

type metricMetaCache struct {
	lastTimestamp int64
	metric        *metrics.Metric
}

type metaProcessor struct {
	metrics   chan *metrics.Metric
	metaCache map[string]*metricMetaCache

	es      writer.Writer
	index   indexmanager.Index
	counter *prometheus.Counter

	r *rand.Rand
}

func createMetaProcess(es writer.Writer, index indexmanager.Index, counter *prometheus.Counter) *metaProcessor {
	p := &metaProcessor{
		metrics:   make(chan *metrics.Metric, 64),
		metaCache: make(map[string]*metricMetaCache),
		es:        es,
		index:     index,
		counter:   counter,
		r:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	go p.process()
	return p
}

func (p *metaProcessor) add(metric *Metric) error {

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

	if metric.Name == MetricMeta {
		return nil
	}

	if metaVal, ok := metric.Tags[MetricInternalPrefix+MetricTagMetricMeta]; ok {
		if b, err := strconv.ParseBool(metaVal); err == nil && b {
			tags := make([]string, 0)
			fields := make([]string, 0)
			if metric.Tags != nil && len(metric.Tags) > 0 {
				for tag := range metric.Tags {
					// 过滤系统内置的 _ 前缀的tag，如 _id  _meta
					if !strings.HasPrefix(tag, MetricInternalPrefix) {
						tags = append(tags, tag)
					}
				}
			}
			if metric.Fields != nil && len(metric.Fields) > 0 {
				for field, val := range metric.Fields {
					fields = append(fields, field+":"+utils.TypeOf(val))
				}
			}

			metaId := metric.Name + "-" +
				metric.Tags["_"+MetricTagScope] + "-" + metric.Tags["_"+MetricTagScopeID]

			metaTags := map[string]string{}
			if metric.Tags != nil && len(metric.Tags) > 0 {
				for tag, val := range metric.Tags {
					// 过滤系统内置的 _ 前缀的tag，如 _id  _meta
					if strings.HasPrefix(tag, MetricInternalPrefix) {
						metaTags[tag[1:]] = val
					}
				}
			}
			metaTags[MetricTagMetricName] = metric.Name
			metaTags[MetricTagMetricID] = metaId
			metaTags[MetricTagOrgName] = metric.Tags[MetricTagOrgName]
			metaTags[MetricTagClusterName] = metric.Tags[MetricTagClusterName]
			metaMetric := &metrics.Metric{
				Name:      MetricMeta,
				Timestamp: metric.Timestamp,
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
		select {
		case metric, ok := <-p.metrics:
			if ok {
				_ = p.processMetricMeta(metric)
			}
		}
	}
}

func (p *metaProcessor) processMetricMeta(metric *metrics.Metric) error {
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
		//p.L.Infof("push metric meta %s at %d", metric.Name, metaCache.lastTimestamp)
		//countMetric(p.counter, metaCache.metric)
		m := &Metric{}
		m.Metric = metaCache.metric.Copy()
		processTimestampDateFormat(m)
		return p.es.Write(&elasticsearch.Document{Index: p.index.GetWriteFixedIndex(metaCache.metric), ID: metaID, Data: m})
	}
	//return p.output.kafka.Write(meta)
	return nil
}
