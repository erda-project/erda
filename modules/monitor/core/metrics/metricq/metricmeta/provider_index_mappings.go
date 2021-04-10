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
	"fmt"
	"strings"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
	indexmanager "github.com/erda-project/erda/modules/monitor/core/metrics/index"
)

// IndexMappingsGroupProvider .
type IndexMappingsGroupProvider struct {
	index indexmanager.Index
}

// NewIndexMappingsGroupProvider .
func NewIndexMappingsGroupProvider(index indexmanager.Index) (*IndexMappingsGroupProvider, error) {
	return &IndexMappingsGroupProvider{index}, nil
}

// MappingsByID .
func (p *IndexMappingsGroupProvider) MappingsByID(id, scope, scopeID string, names []string, ms map[string]*metrics.MetricMeta) (gmm []*GroupMetricMap, err error) {
	for _, name := range names {
		if mm, ok := ms[name]; ok {
			gmm = append(gmm, &GroupMetricMap{
				Name:            mm.Name.Key,
				AddTagsToFields: true,
			})
		}
	}
	return gmm, nil
}

// Groups .
func (p *IndexMappingsGroupProvider) Groups(langCodes i18n.LanguageCodes, t i18n.Translator, scope, scopeID string, ms map[string]*metrics.MetricMeta) (groups []*Group, err error) {
	names := p.index.MetricNames()
	for _, name := range names {
		if strings.HasPrefix(name, "_") {
			continue
		}
		groups = append(groups, &Group{
			ID:   name + "@" + name,
			Name: t.Text(langCodes, name),
		})
	}
	return groups, nil
}

// IndexMappingsMetricMetaProvider .
type IndexMappingsMetricMetaProvider struct {
	index indexmanager.Index
}

// NewIndexMappingsMetricMetaProvider .
func NewIndexMappingsMetricMetaProvider(index indexmanager.Index) (*IndexMappingsMetricMetaProvider, error) {
	return &IndexMappingsMetricMetaProvider{index}, nil
}

// MetricMeta .
func (p *IndexMappingsMetricMetaProvider) MetricMeta(langCodes i18n.LanguageCodes, i i18n.I18n, scope, scopeID string, names ...string) (map[string]*metrics.MetricMeta, error) {
	var indices []string
	if len(names) > 0 {
		for _, n := range names {
			indices = append(indices, p.index.IndexPrefix()+"-"+n+"-*")
		}
	} else {
		indices = append(indices, p.index.IndexPrefix()+"-*")
	}
	resp, err := p.getMappings(indices)
	if err != nil {
		return nil, err
	}
	metas := make(map[string]*metrics.MetricMeta)
	for key, m := range resp {
		if len(key) < len(p.index.IndexPrefix())+1 {
			continue
		}
		keys := strings.Split(key, "-")
		key = keys[1]
		if strings.HasPrefix(key, "_") || key == "empty" {
			continue
		}
		mp, ok := m.(map[string]interface{})
		if !ok {
			continue
		}
		mp, ok = mp["mappings"].(map[string]interface{})
		if !ok {
			continue
		}
		if len(mp) <= 0 {
			continue
		}
		for _, m := range mp {
			m, ok := m.(map[string]interface{})
			if ok {
				mp = m
				break
			}
		}
		if len(mp) <= 0 {
			continue
		}
		mp, ok = mp["properties"].(map[string]interface{})
		if !ok {
			continue
		}
		tags, ok := mp["tags"].(map[string]interface{})
		if !ok {
			continue
		}
		tags, ok = tags["properties"].(map[string]interface{})
		if !ok {
			continue
		}
		fields, ok := mp["fields"].(map[string]interface{})
		if !ok {
			continue
		}
		fields, ok = fields["properties"].(map[string]interface{})
		if !ok {
			continue
		}
		meta := metrics.NewMeta()
		meta.Name.Key, meta.Name.Name = key, key
		for k := range tags {
			if !strings.HasPrefix(k, "_") {
				meta.Tags[k] = &metrics.TagDefine{
					Key:  k,
					Name: k,
				}
			}
		}
		for k, v := range fields {
			info, ok := v.(map[string]interface{})
			if !ok {
				continue
			}
			typ, ok := info["type"]
			if !ok {
				continue
			}
			t := fmt.Sprint(typ)
			switch t {
			case "keyword", "string":
				t = StringType
			case "double", "long":
				t = NumberType
			case "boolean":
				t = BoolType
			default:
				t = StringType
			}
			meta.Fields[k] = &metrics.FieldDefine{
				Key:  k,
				Type: t,
				Name: k,
				Unit: "",
			}
		}
		metas[key] = mergeMetricMeta(meta, metas[key])
	}
	return transMetricMetas(langCodes, i, metas), nil
}

func mergeMetricMeta(m1, m2 *metrics.MetricMeta) *metrics.MetricMeta {
	if m2 == nil {
		return m1
	}
	if m1 == nil {
		return m2
	}
	for t, tag := range m2.Tags {
		if _, ok := m1.Tags[t]; !ok {
			m1.Tags[t] = tag
		}
	}
	for f, field := range m2.Fields {
		if _, ok := m1.Fields[f]; !ok {
			m1.Fields[f] = field
		}
	}
	return m1
}

func (p *IndexMappingsMetricMetaProvider) getMappings(indices []string) (map[string]interface{}, error) {
	context, cancel := context.WithTimeout(context.Background(), p.index.RequestTimeout()*2)
	defer cancel()
	return p.index.Client().GetMapping().Index(indices...).IgnoreUnavailable(true).AllowNoIndices(true).Do(context)
}
