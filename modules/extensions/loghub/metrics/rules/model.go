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

package rules

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/erda-project/erda/modules/extensions/loghub/metrics/rules/db"
	"github.com/recallsong/go-utils/reflectx"
)

// LogMetricConfig .
type LogMetricConfig struct {
	ID         int64              `json:"id"`
	OrgID      int64              `json:"org_id"`
	Scope      string             `json:"scope"`
	ScopeID    string             `json:"Scope_id"`
	Name       string             `json:"name"`
	Metric     string             `json:"metric"`
	Filters    []*Tag             `json:"filters"`
	Processors []*ProcessorConfig `json:"processors"`
	Enable     bool               `json:"enable"`
	CreateTime int64              `json:"create_time"`
	UpdateTime int64              `json:"update_time"`
}

// Tag .
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ProcessorConfig .
type ProcessorConfig struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// FromModel .
func (c *LogMetricConfig) FromModel(m *db.LogMetricConfig) *LogMetricConfig {
	if len(m.Filters) > 0 {
		var tags []*Tag
		err := json.Unmarshal([]byte(m.Filters), &tags)
		if err == nil {
			c.Filters = tags
		}
	}
	if len(m.Processors) > 0 {
		var ps []*ProcessorConfig
		err := json.Unmarshal([]byte(m.Processors), &ps)
		if err == nil {
			c.Processors = ps
		}
	}
	c.ID = m.ID
	c.OrgID = m.OrgID
	c.Scope = m.Scope
	c.ScopeID = m.ScopeID
	c.Name = m.Name
	c.Metric = m.Metric
	c.Enable = m.Enable
	c.CreateTime = m.CreateTime.UnixNano() / int64(time.Millisecond)
	c.UpdateTime = m.UpdateTime.UnixNano() / int64(time.Millisecond)
	return c
}

// ToModel .
func (c *LogMetricConfig) ToModel() *db.LogMetricConfig {
	filters, _ := json.Marshal(c.Filters)
	processors, _ := json.Marshal(c.Processors)
	return &db.LogMetricConfig{
		ID:         c.ID,
		OrgID:      c.OrgID,
		Scope:      c.Scope,
		ScopeID:    c.ScopeID,
		Name:       c.Name,
		Metric:     c.Metric,
		Filters:    string(filters),
		Processors: string(processors),
		Enable:     c.Enable,
		CreateTime: time.Unix(c.CreateTime/1000, (c.CreateTime%1000)*int64(time.Millisecond)),
		UpdateTime: time.Unix(c.UpdateTime/1000, (c.UpdateTime%1000)*int64(time.Millisecond)),
	}
}

// LogMetricConfigSimple .
type LogMetricConfigSimple struct {
	ID         int64  `json:"id"`
	OrgID      int64  `json:"org_id"`
	Scope      string `json:"scope"`
	ScopeID    string `json:"scope_id"`
	Name       string `json:"name"`
	Types      string `json:"types"`
	Metric     string `json:"metric"`
	Enable     bool   `json:"enable"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

// FromModel .
func (c *LogMetricConfigSimple) FromModel(m *db.LogMetricConfig) *LogMetricConfigSimple {
	c.ID = m.ID
	c.OrgID = m.OrgID
	c.Scope = m.Scope
	c.ScopeID = m.ScopeID
	c.Name = m.Name
	c.Metric = m.Metric
	c.Enable = m.Enable
	c.CreateTime = m.CreateTime.UnixNano() / int64(time.Millisecond)
	c.UpdateTime = m.UpdateTime.UnixNano() / int64(time.Millisecond)
	// ProcessorConfig .
	type ProcessorConfig struct {
		Type string `json:"type"`
	}
	var processors []*ProcessorConfig
	err := json.Unmarshal(reflectx.StringToBytes(m.Processors), &processors)
	if err == nil {
		var typs []string
		for _, p := range processors {
			typs = append(typs, p.Type)
		}
		c.Types = strings.Join(typs, ",")
	}
	return c
}
