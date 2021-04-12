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

package metricmeta

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
	indexmanager "github.com/erda-project/erda/modules/monitor/core/metrics/index"
	"github.com/olivere/elastic"
)

// MetaIndexGroupProvider .
type MetaIndexGroupProvider struct {
	index indexmanager.Index
}

// NewMetaIndexGroupProvider .
func NewMetaIndexGroupProvider(index indexmanager.Index) (*MetaIndexGroupProvider, error) {
	return &MetaIndexGroupProvider{index}, nil
}

// MappingsByID .
func (p *MetaIndexGroupProvider) MappingsByID(id, scope, scopeID string, names []string, ms map[string]*metrics.MetricMeta) (gmm []*GroupMetricMap, err error) {
	if id == "custom" || id == "other" {
		for _, name := range names {
			if mm, ok := ms[name]; ok {
				gmm = append(gmm, &GroupMetricMap{
					Name: mm.Name.Key,
				})
			}
		}
	}
	return gmm, nil
}

// Groups .
func (p *MetaIndexGroupProvider) Groups(langCodes i18n.LanguageCodes, t i18n.Translator, scope, scopeID string, ms map[string]*metrics.MetricMeta) (groups []*Group, err error) {
	if scope == "org" || scope == "dice" { // 暂时硬编码
		groups = append(groups, &Group{
			ID:    "other",
			Name:  t.Text(langCodes, "Other"),
			Order: math.MaxInt32,
		})
		groups = appendMetricToGroup(groups, "@", ms, p.getOtherGroupsMetrics(ms), true)
	} else {
		groups = append(groups, &Group{
			ID:    "custom",
			Name:  t.Text(langCodes, "Custom"),
			Order: math.MaxInt32,
		})
		groups = appendMetricToGroup(groups, "@", ms, p.getCustomGroupsMetrics(ms), true)
	}
	return groups, nil
}

func (p *MetaIndexGroupProvider) getCustomGroupsMetrics(ms map[string]*metrics.MetricMeta) map[string][]*GroupMetricMap {
	return p.getDynamicGroupsMetrics("custom", ms, func(m *metrics.MetricMeta) bool {
		return m.Labels != nil && m.Labels["custom"] == "true"
	})
}

func (p *MetaIndexGroupProvider) getOtherGroupsMetrics(ms map[string]*metrics.MetricMeta) map[string][]*GroupMetricMap {
	return p.getDynamicGroupsMetrics("other", ms, func(m *metrics.MetricMeta) bool {
		return true
	})
}

func (p *MetaIndexGroupProvider) getDynamicGroupsMetrics(group string, ms map[string]*metrics.MetricMeta, match func(m *metrics.MetricMeta) bool) map[string][]*GroupMetricMap {
	var gm map[string][]*GroupMetricMap
	for _, m := range ms {
		if m == nil {
			continue
		}
		if match(m) {
			if gm == nil {
				gm = make(map[string][]*GroupMetricMap)
			}
			gmm := &GroupMetricMap{
				Name:   m.Name.Key,
				Fields: m.FieldsKeys(),
			}
			gm[group] = append(gm[group], gmm)
		}
	}
	for _, list := range gm {
		if len(list) == 0 {
			continue
		}
		sort.Slice(list, func(i, j int) bool {
			if list[i].Name == list[j].Name {
				return len(list[i].Fields) < len(list[j].Fields)
			}
			return list[i].Name < list[j].Name
		})
	}
	return gm
}

// MetaIndexMetricMetaProvider .
type MetaIndexMetricMetaProvider struct {
	index indexmanager.Index
	log   logs.Logger
}

// NewMetaIndexMetricMetaProvider .
func NewMetaIndexMetricMetaProvider(index indexmanager.Index, log logs.Logger) (*MetaIndexMetricMetaProvider, error) {
	return &MetaIndexMetricMetaProvider{
		index: index,
		log:   log,
	}, nil
}

// MetricMeta .
func (p *MetaIndexMetricMetaProvider) MetricMeta(langCodes i18n.LanguageCodes, i i18n.I18n, scope, scopeID string, names ...string) (map[string]*metrics.MetricMeta, error) {
	query := elastic.NewBoolQuery().
		Filter(elastic.NewExistsQuery("fields.fields")).
		Filter(elastic.NewExistsQuery("fields.tags"))
	if scope != "org" && scope != "dice" { // 暂时硬编码
		query = query.Filter(elastic.NewTermQuery("tags.metric_scope", scope)).
			Filter(elastic.NewTermQuery("tags.metric_scope_id", scopeID))
	}
	if len(names) > 0 {
		var args []interface{}
		for _, item := range names {
			args = append(args, item)
		}
		query = query.Filter(elastic.NewTermsQuery("tags.metric_name", args...))
	}

	searchSource := elastic.NewSearchSource().
		Query(query).Size(0).
		Aggregation("tags.metric_name", elastic.NewTermsAggregation().Field("tags.metric_name").Size(100000). //  impossible size
															SubAggregation("topHit", elastic.NewTopHitsAggregation().Size(1).Sort("timestamp", false).
																FetchSourceContext(elastic.NewFetchSourceContext(true).Include("*"))))
	end := time.Now().UnixNano() / int64(time.Millisecond)
	start := end - 7*24*int64(time.Hour)/int64(time.Millisecond)
	indices := p.index.GetReadIndices([]string{"_metric_meta"}, nil, start, end)
	result, err := p.searchRaw(indices, searchSource)
	if err != nil {
		return nil, err
	}
	metas := make(map[string]*metrics.MetricMeta)
	if result == nil || result.Aggregations == nil {
		return metas, nil
	}
	metricNameAgg, ok := result.Aggregations.Terms("tags.metric_name")
	if !ok {
		return metas, nil
	}

	type ESMetricMeta struct {
		Name      string `json:"name"`
		Timestamp int64  `json:"timestamp"`
		Fields    *struct {
			FieldsKeys []string `json:"fields"`
			TagsKeys   []string `json:"tags"`
		} `json:"fields"`
		Tags map[string]string `json:"tags"`
	}
	for _, item := range metricNameAgg.Buckets {
		var data ESMetricMeta
		topHitAgg, _ := item.TopHits("topHit")
		if topHitAgg == nil || topHitAgg.Hits == nil || len(topHitAgg.Hits.Hits) == 0 {
			continue
		}
		hit := topHitAgg.Hits.Hits[0]
		if hit.Source == nil {
			continue
		}
		if err := json.Unmarshal(*hit.Source, &data); err != nil {
			p.log.Warnf("fail to json decode metric meta %s: %s", hit.Source, err)
			continue
		}
		if data.Tags == nil || data.Fields == nil {
			continue
		}
		metricName := data.Tags["metric_name"]
		if len(metricName) <= 0 {
			continue
		}
		labels := make(map[string]string)
		for k, v := range data.Tags {
			labels[k] = v
		}
		meta := metrics.NewMeta()
		meta.Name.Key, meta.Name.Name = metricName, metricName
		if len(labels) > 0 {
			meta.Labels = labels
		}
		for _, key := range data.Fields.TagsKeys {
			meta.Tags[key] = &metrics.TagDefine{
				Key:  key,
				Name: key,
			}
		}
		for _, key := range data.Fields.FieldsKeys {
			idx := strings.LastIndex(key, ":")
			if idx <= 0 {
				p.log.Warnf("invalid field format in %s: %s", key, metricName)
				continue
			}
			meta.Fields[key[:idx]] = &metrics.FieldDefine{
				Key:  key[:idx],
				Type: key[idx+1:],
				Name: key[:idx],
				Unit: "",
			}
		}
		metas[metricName] = meta
	}
	return transMetricMetas(langCodes, i, metas), nil
}

func (p *MetaIndexMetricMetaProvider) searchRaw(indices []string, searchSource *elastic.SearchSource) (*elastic.SearchResult, error) {
	context, cancel := context.WithTimeout(context.Background(), p.index.RequestTimeout())
	defer cancel()
	return p.index.Client().Search(indices...).
		IgnoreUnavailable(true).AllowNoIndices(true).
		SearchSource(searchSource).Do(context)
}
