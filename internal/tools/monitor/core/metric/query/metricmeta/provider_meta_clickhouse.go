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

package metricmeta

import (
	"context"
	"math"
	"sort"

	"github.com/erda-project/erda-infra/providers/i18n"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse/table/meta"
)

type MetaClickhouseGroupProvider struct {
	ckMetaLoader meta.Interface
}

func NewMetaClickhouseGroupProvider(ck meta.Interface) (*MetaClickhouseGroupProvider, error) {
	return &MetaClickhouseGroupProvider{
		ckMetaLoader: ck,
	}, nil
}

func (m *MetaClickhouseGroupProvider) MappingsByID(id, scope, scopeID string, names []string, ms map[string]*metricpb.MetricMeta) (gmm []*GroupMetricMap, err error) {
	if id == "all" {
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

func (m *MetaClickhouseGroupProvider) Groups(langCodes i18n.LanguageCodes, t i18n.Translator, scope, scopeID string, ms map[string]*metricpb.MetricMeta) (groups []*metricpb.Group, err error) {
	groups = append(groups, &metricpb.Group{
		Id:    "all",
		Name:  t.Text(langCodes, "All Metrics"),
		Order: math.MaxInt32,
	})
	groups = appendMetricToGroup(groups, "@", ms, m.getAllGroupsMetrics(ms), true)
	return groups, nil
}
func (m *MetaClickhouseGroupProvider) getAllGroupsMetrics(ms map[string]*metricpb.MetricMeta) map[string][]*GroupMetricMap {
	return m.getDynamicGroupsMetrics("all", ms, func(m *metricpb.MetricMeta) bool {
		return true
	})
}

func (m *MetaClickhouseGroupProvider) getDynamicGroupsMetrics(group string, ms map[string]*metricpb.MetricMeta, match func(m *metricpb.MetricMeta) bool) map[string][]*GroupMetricMap {
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
				Fields: metric.FieldsKeys(m),
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

func (p MetaClickhouseGroupProvider) MetricMeta(langCodes i18n.LanguageCodes, i i18n.I18n, scope, scopeID string, names ...string) (map[string]*metricpb.MetricMeta, error) {
	result := p.ckMetaLoader.GetMeta(context.Background(), scope, scopeID, names...)
	if len(result) <= 0 {
		return map[string]*metricpb.MetricMeta{}, nil
	}
	metas := make(map[string]*metricpb.MetricMeta)
	for _, cm := range result {
		meta := metric.NewMeta()
		meta.Name.Key, meta.Name.Name = cm.MetricGroup, cm.MetricGroup
		meta.Tags = make(map[string]*metricpb.TagDefine)
		for _, tag := range cm.TagKeys {
			meta.Tags[tag] = &metricpb.TagDefine{
				Key:  tag,
				Name: tag,
			}
		}
		meta.Fields = make(map[string]*metricpb.FieldDefine)
		for _, field := range cm.NumberKeys {
			meta.Fields[field] = &metricpb.FieldDefine{
				Key:  field,
				Name: field,
				Type: "number",
			}
		}
		for _, field := range cm.StringKeys {
			meta.Fields[field] = &metricpb.FieldDefine{
				Key:  field,
				Name: field,
				Type: "string",
			}
		}
		metas[cm.MetricGroup] = meta
	}
	return transMetricMetas(langCodes, i, metas), nil
}
