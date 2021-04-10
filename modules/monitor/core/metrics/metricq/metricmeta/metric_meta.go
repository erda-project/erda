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
	"fmt"
	"sort"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
)

type MetricMetaProvider interface {
	MetricMeta(langCodes i18n.LanguageCodes, i i18n.I18n, scope, scopeID string, names ...string) (map[string]*metrics.MetricMeta, error)
}

func (m *Manager) getMetricMetaProviders() (list []MetricMetaProvider) {
	for _, fn := range m.metricProviders {
		list = append(list, fn())
	}
	return list
}

func (m *Manager) MetricNames(langCodes i18n.LanguageCodes, scope, scopeID string) (names []*metrics.NameDefine, err error) {
	metrics, err := m.MetricMeta(langCodes, scope, scopeID)
	if err != nil {
		return nil, err
	}
	for _, m := range metrics {
		names = append(names, &m.Name)
	}
	return names, nil
}

func (m *Manager) MetricMeta(langCodes i18n.LanguageCodes, scope, scopeID string, names ...string) ([]*metrics.MetricMeta, error) {
	metricMetas, err := m.getMetricMeta(langCodes, scope, scopeID, names...)
	if err != nil {
		return nil, err
	}
	var list []*metrics.MetricMeta
	for _, item := range metricMetas {
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].Name.Name == list[j].Name.Name {
			return list[i].Name.Key == list[j].Name.Key
		}
		return list[i].Name.Name < list[j].Name.Name
	})
	return list, nil
}

func (m *Manager) GetSingleMetricsMeta(langCodes i18n.LanguageCodes, scope string, scopeID string, metric string) (*metrics.MetricMeta, error) {
	metricMetas, err := m.getMetricMeta(langCodes, scope, scopeID, metric)
	if err != nil {
		return nil, err
	}
	metricMeta, ok := metricMetas[metric]
	if !ok {
		return nil, fmt.Errorf("can't find metric: %s", metric)
	}
	return metricMeta, nil
}

func (m *Manager) getMetricMeta(langCodes i18n.LanguageCodes, scope, scopeID string, names ...string) (map[string]*metrics.MetricMeta, error) {
	mp := m.getMetricMetaProviders()
	ms := make(map[string]*metrics.MetricMeta)
	for _, p := range mp {
		m, err := p.MetricMeta(langCodes, m.i18n, scope, scopeID, names...)
		if err != nil {
			return nil, err
		}
		ms = appendMetricMeta(ms, m)
	}
	return ms, nil
}

func appendMetricMeta(metric1, metric2 map[string]*metrics.MetricMeta) map[string]*metrics.MetricMeta {
	for n1, m1 := range metric1 {
		if m2, ok := metric2[n1]; ok {
			m1.Name = m2.Name
			m1.Labels = appendLabels(m1.Labels, m2.Labels)
			m1.Fields = appendFields(m1.Fields, m2.Fields)
			m1.Tags = appendTags(m1.Tags, m2.Tags)
		}
	}
	for n2, m2 := range metric2 {
		if _, ok := metric1[n2]; !ok {
			metric1[n2] = m2
		}
	}
	return metric1
}

func appendTags(a, b map[string]*metrics.TagDefine) map[string]*metrics.TagDefine {
	if a == nil {
		return b
	}
	if b != nil {
		for k, v := range b {
			// if _, ok := a[k]; ok {
			// 	continue
			// }
			a[k] = v
		}
	}
	return a
}

func appendFields(a, b map[string]*metrics.FieldDefine) map[string]*metrics.FieldDefine {
	if a == nil {
		return b
	}
	if b != nil {
		for k, v := range b {
			// if _, ok := a[k]; ok {
			// 	continue
			// }
			a[k] = v
		}
	}
	return a
}

func appendLabels(a, b map[string]string) map[string]string {
	if a == nil {
		return b
	}
	if b != nil {
		for k, v := range b {
			// if _, ok := a[k]; ok {
			// 	continue
			// }
			a[k] = v
		}
	}
	return a
}

func copyMetricMeta(m *metrics.MetricMeta) *metrics.MetricMeta {
	n := metrics.NewMeta()
	n.Name = m.Name
	for k, t := range m.Tags {
		tag := *t
		var values []*metrics.ValueDefine
		for _, v := range t.Values {
			nv := *v
			values = append(values, &nv)
		}
		tag.Values = values
		n.Tags[k] = &tag
	}
	for k, f := range m.Fields {
		field := *f
		var values []*metrics.ValueDefine
		for _, v := range f.Values {
			nv := *v
			values = append(values, &nv)
		}
		field.Values = values
		n.Fields[k] = &field
	}
	return n
}

func transMetricMetas(langCodes i18n.LanguageCodes, i i18n.I18n, metas map[string]*metrics.MetricMeta) map[string]*metrics.MetricMeta {
	for _, item := range metas {
		t := i.Translator(item.Name.Key)
		item.Name.Name = t.Text(langCodes, item.Name.Name)
		for _, f := range item.Fields {
			f.Name = t.Text(langCodes, f.Name)
			for _, v := range f.Values {
				v.Name = t.Text(langCodes, v.Name)
			}
		}
		for _, tag := range item.Tags {
			tag.Name = t.Text(langCodes, tag.Name)
			for _, v := range tag.Values {
				v.Name = t.Text(langCodes, v.Name)
			}
		}
	}
	return metas
}

func (m *Manager) RegeistMetricMeta(scope, scopeID, group string, metrics ...*metrics.MetricMeta) error {
	return m.regeistMetricMeta(scope, scopeID, group, metrics...)
}

func (m *Manager) UnregeistMetricMeta(scope, scopeID, group string, metrics ...string) error {
	return m.unregeistMetricMeta(scope, scopeID, group, metrics...)
}
