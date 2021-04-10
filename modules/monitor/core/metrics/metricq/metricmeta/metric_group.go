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
	"sort"
	"strings"

	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
)

type Group struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Order    int32    `json:"-"`
	Children []*Group `json:"children,omitempty"`
}

type Filter struct {
	Tag   string `json:"tag"`
	Op    string `json:"op"`
	Value string `json:"value"`
}

type GroupMetricMap struct {
	Name            string    `json:"name"`
	Filters         []*Filter `json:"filters"`
	Fields          []string  `json:"fields"`
	Tags            []string  `json:"tags"`
	AddTagsToFields bool      `json:"add_tags_to_fields"`
}

type GroupProvider interface {
	MappingsByID(id, scope, scopeID string, metric []string, ms map[string]*metrics.MetricMeta) ([]*GroupMetricMap, error)
	Groups(langCodes i18n.LanguageCodes, t i18n.Translator, scope, scopeID string, ms map[string]*metrics.MetricMeta) ([]*Group, error)
}

func (m *Manager) getGroupProviers() (list []GroupProvider) {
	for _, fn := range m.groupProviders {
		list = append(list, fn())
	}
	return list
}

func (m *Manager) getGroupAndMetric(id string) (string, []string) {
	// for _, sep := range []string{"@", "#"} {
	// idx := strings.Index(id, sep)
	idx := strings.Index(id, "@")
	if idx > 0 {
		if len(id[idx+1:]) > 0 {
			return id[:idx], []string{id[idx+1:]}
		}
		return id[:idx], nil
	}
	//}
	return id, nil
}

func (m *Manager) MetricGroups(langCodes i18n.LanguageCodes, scope, scopeID, mode string) ([]*Group, error) {
	ms, err := m.getMetricMeta(langCodes, scope, scopeID)
	if err != nil {
		return nil, err
	}
	gp := m.getGroupProviers()
	var gs []*Group
	t := m.i18n.Translator("_group")
	for _, p := range gp {
		g, err := p.Groups(langCodes, t, scope, scopeID, ms)
		if err != nil {
			return nil, err
		}
		gs = appendGroups(gs, g)
	}
	sortGroups(gs)
	return gs, nil
}

func appendGroups(group1, group2 []*Group) []*Group {
	for _, g1 := range group1 {
		for _, g2 := range group2 {
			if g1.ID == g2.ID {
				if len(g2.Name) > 0 {
					g1.Name = g2.Name
				}
				g1.Children = appendGroups(g1.Children, g2.Children)
				break
			}
		}
	}
	for _, g2 := range group2 {
		var find bool
		for _, g1 := range group1 {
			if g2.ID == g1.ID {
				find = true
				break
			}
		}
		if !find {
			group1 = append(group1, g2)
		}
	}
	return group1
}

func sortGroups(groups []*Group) {
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Order < groups[j].Order
	})
	for _, g := range groups {
		if len(g.Children) <= 0 {
			continue
		}
		sortGroups(g.Children)
	}
}

// formats .
const (
	InfluxFormat = "influx"
	DotFormat    = "dot"
)

type GroupDetail struct {
	ID      string             `json:"id"`
	Meta    *MetaMode          `json:"meta"`
	Metrics []*GroupMetricMeta `json:"metrics"`
}

type GroupMetricMeta struct {
	Metric  string                 `json:"metric"`
	Name    string                 `json:"name"`
	Filters []*Filter              `json:"filters,omitempty"`
	Fields  []*metrics.FieldDefine `json:"fields"`
	Tags    []*metrics.TagDefine   `json:"tags"`
}

func (m *Manager) MetricGroup(langCodes i18n.LanguageCodes, scope, scopeID, id, mode, format string, appendTags bool) (*GroupDetail, error) {
	groupID, metric := m.getGroupAndMetric(id)
	ms, err := m.getMetricMeta(langCodes, scope, scopeID, metric...)
	if err != nil {
		return nil, err
	}
	gp := m.getGroupProviers()
	var mappings []*GroupMetricMap
	for _, p := range gp {
		m, err := p.MappingsByID(groupID, scope, scopeID, metric, ms)
		if err != nil {
			return nil, err
		}
		mappings = append(mappings, m...)
	}

	gd := &GroupDetail{
		ID: id,
	}
	if format == InfluxFormat {
		gd.Meta, err = m.getTypeAggDefineInflux(langCodes, mode)
		if err != nil {
			return nil, err
		}
	} else {
		gd.Meta, err = m.getTypeAggDefine(langCodes, mode)
		if err != nil {
			return nil, err
		}
	}

	for _, gm := range mappings {
		mm := ms[gm.Name]
		if mm == nil {
			continue
		}
		gmm := &GroupMetricMeta{
			Metric:  mm.Name.Key,
			Name:    mm.Name.Name,
			Filters: gm.Filters,
		}

		if len(gm.Fields) <= 0 {
			for _, f := range mm.FieldsKeys() {
				fd := mm.Fields[f]
				if fd != nil {
					fd.Key = getFieldKeyWithFormat(format, fd.Key, "field")
					gmm.Fields = append(gmm.Fields, fd)
				}
			}
		} else {
			for _, f := range gm.Fields {
				fd := mm.Fields[f]
				if fd != nil {
					fd.Key = getFieldKeyWithFormat(format, fd.Key, "field")
					gmm.Fields = append(gmm.Fields, fd)
				}
			}
		}
		if len(gm.Tags) > 0 {
			for _, t := range gm.Tags {
				td := mm.Tags[t]
				if td != nil {
					fd := &metrics.FieldDefine{
						Key:    getFieldKeyWithFormat(format, t, "tag"),
						Type:   StringType,
						Name:   td.Name,
						Values: td.Values,
					}
					gmm.Fields = append(gmm.Fields, fd)
				}
			}
		} else if gm.AddTagsToFields || appendTags {
			for _, t := range mm.TagsKeys() {
				td := mm.Tags[t]
				if td != nil {
					fd := &metrics.FieldDefine{
						Key:    getFieldKeyWithFormat(format, t, "tag"),
						Type:   StringType,
						Name:   td.Name,
						Values: td.Values,
					}
					gmm.Fields = append(gmm.Fields, fd)
				}
			}
		}
		for _, t := range mm.TagsKeys() {
			td := mm.Tags[t]
			if td != nil {
				gmm.Tags = append(gmm.Tags, mm.Tags[t])
			}
		}
		gd.Metrics = append(gd.Metrics, gmm)
	}
	return gd, nil
}

func getFieldKeyWithFormat(format, key, typ string) string {
	if format == InfluxFormat {
		return key + "::" + typ
	} else if format == DotFormat || typ == "tag" {
		return typ + "s." + key
	}
	return key
}

func appendMetricToGroup(groups []*Group, sep string, metricmeta map[string]*metrics.MetricMeta, mapping map[string][]*GroupMetricMap, forceAppend bool) []*Group {
	for _, g := range groups {
		g.Children = appendMetricToGroup(g.Children, sep, metricmeta, mapping, forceAppend)
		if len(mapping[g.ID]) > 1 || forceAppend {
			for _, m := range mapping[g.ID] {
				c := &Group{
					ID:   g.ID + sep + m.Name,
					Name: m.Name,
				}
				meta := metricmeta[m.Name]
				if meta != nil {
					c.Name = meta.Name.Name
				}
				g.Children = append(g.Children, c)
			}
		}
	}
	return groups
}
