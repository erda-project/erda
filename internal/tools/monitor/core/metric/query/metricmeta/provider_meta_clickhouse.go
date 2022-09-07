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
	"math"
	"sort"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/pkg/errors"

	"github.com/erda-project/erda-infra/providers/i18n"
	metricpb "github.com/erda-project/erda-proto-go/core/monitor/metric/pb"
	"github.com/erda-project/erda/internal/tools/monitor/core/metric"
	"github.com/erda-project/erda/internal/tools/monitor/core/storekit/clickhouse"
)

type MetaClickhouseGroupProvider struct {
	clickhouse clickhouse.Query
}

func NewMetaClickhouseGroupProvider(ck clickhouse.Query) (*MetaClickhouseGroupProvider, error) {
	return &MetaClickhouseGroupProvider{
		clickhouse: ck,
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

type ckMeta struct {
	MetricGroup string   `ch:"metric_group"`
	StringKeys  []string `ch:"sk"`
	NumberKeys  []string `ch:"nk"`
	TagKeys     []string `ch:"tk"`
}

var now = func() time.Time {
	return time.Now()
}

func (p MetaClickhouseGroupProvider) MetricMeta(langCodes i18n.LanguageCodes, i i18n.I18n, scope, scopeID string, names ...string) (map[string]*metricpb.MetricMeta, error) {
	if p.clickhouse == nil {
		return map[string]*metricpb.MetricMeta{}, nil
	}

	/*
		CREATE TABLE IF NOT EXISTS <database>.metrics_meta ON CLUSTER '{cluster}'
		(
		    `org_name`            LowCardinality(String),
		    `tenant_id`           LowCardinality(String),
		    `metric_group`        LowCardinality(String),
		    `timestamp`           DateTime64(9,'Asia/Shanghai') CODEC (DoubleDelta),
		    `number_field_keys`   Array(LowCardinality(String)),
		    `string_field_keys`   Array(LowCardinality(String)),
		    `tag_keys`            Array(LowCardinality(String)),
		    INDEX idx_timestamp TYPE minmax GRANULARITY 2
		)
		ENGINE = ReplicatedReplacingMergeTree('/clickhouse/tables/{cluster}-{shard}/{database}/metrics_meta', '{replica}')
		ORDER BY (org_name, tenant_id, metric_group, number_field_keys, string_field_keys, tag_keys);
		TTL toDateTime(timestamp) + INTERVAL <ttl_in_days> DAY;
	*/

	end := now().UnixNano()
	start := end - 7*24*int64(time.Hour)

	expr := goqu.From("metrics_meta")

	expr = goqu.Select(goqu.C("metric_group"))
	expr = expr.SelectAppend(goqu.L("string_field_keys").As("sk"))
	expr = expr.SelectAppend(goqu.L("number_field_keys").As("nk"))
	expr = expr.SelectAppend(goqu.L("tag_keys").As("tk"))

	expr = expr.Where(goqu.C("org_name").Eq(scope))
	if len(scopeID) > 0 {
		expr = expr.Where(goqu.C("tenant_id").Eq(scopeID))
	}
	if len(names) > 0 {
		expr = expr.Where(goqu.C("metric_group").In(names))
	}

	expr = expr.Where(
		goqu.C("timestamp").Gte(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", start)),
		goqu.C("timestamp").Lt(goqu.L("fromUnixTimestamp64Nano(cast(?,'Int64'))", end)),
	)
	expr = expr.GroupBy(goqu.C("metric_group"))

	rows, err := p.clickhouse.QueryRaw(scope, expr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query metric meta")
	}

	metas := make(map[string]*metricpb.MetricMeta)

	for rows.Next() {
		var cm ckMeta
		err := rows.ScanStruct(&cm)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan metric meta")
		}
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
