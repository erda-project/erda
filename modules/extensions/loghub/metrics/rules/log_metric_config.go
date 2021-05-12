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

	"github.com/recallsong/go-utils/encoding/md5x"
	uuid "github.com/satori/go.uuid"

	"github.com/erda-project/erda/modules/extensions/loghub/metrics/analysis/processors"
	"github.com/erda-project/erda/modules/extensions/loghub/metrics/rules/db"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
	"github.com/erda-project/erda/modules/pkg/mysql"
)

// ListLogMetricConfig .
func (p *provider) ListLogMetricConfig(scope, scopeID string) ([]*LogMetricConfigSimple, error) {
	var result []*LogMetricConfigSimple
	list, err := p.db.LogMetricConfig.QueryByScope(scope, scopeID)
	if err != nil {
		return nil, err
	}
	for _, item := range list {
		result = append(result, (&LogMetricConfigSimple{}).FromModel(item))
	}
	return result, nil
}

// GetLogMetricConfig .
func (p *provider) GetLogMetricConfig(scope, scopeID string, id int64) (*LogMetricConfig, error) {
	c, err := p.db.LogMetricConfig.QueryByID(scope, scopeID, id)
	if err != nil {
		return nil, err
	}
	cfg := (&LogMetricConfig{}).FromModel(c)
	return cfg, nil
}

func (p *provider) makeIndex() string {
	return "log_" + md5x.SumString(uuid.NewV4().String()).String16()
}

// CreateLogMetricConfig .
func (p *provider) CreateLogMetricConfig(cfg *LogMetricConfig) (bool, error) {
	m := cfg.ToModel()
	m.Metric = p.makeIndex()
	meta, err := p.convertToMetricMeta(m)
	if err != nil {
		return false, err
	}
	err = p.db.LogMetricConfig.Insert(m)
	if err != nil {
		if mysql.IsUniqueConstraintError(err) {
			return true, err
		}
		return false, err
	}
	err = p.metricq.RegeistMetricMeta(cfg.Scope, cfg.ScopeID, "log_metrics", meta)
	if err != nil {
		return false, err
	}
	return false, nil
}

// EnableLogMetricConfig .
func (p *provider) EnableLogMetricConfig(scope, scopeID string, id int64, enable bool) error {
	return p.db.LogMetricConfig.Enable(scope, scopeID, id, enable)
}

// UpdateLogMetricConfig .
func (p *provider) UpdateLogMetricConfig(cfg *LogMetricConfig) (bool, error) {
	db := p.db.Begin()
	err := db.LogMetricConfig.Update(cfg.ToModel())
	if err != nil {
		db.Rollback()
		if mysql.IsUniqueConstraintError(err) {
			return true, err
		}
		return false, err
	}

	m, err := db.LogMetricConfig.QueryByID(cfg.Scope, cfg.ScopeID, cfg.ID)
	if err != nil {
		db.Rollback()
		return false, err
	}
	meta, err := p.convertToMetricMeta(m)
	if err != nil {
		db.Rollback()
		return false, err
	}
	err = p.metricq.RegeistMetricMeta(cfg.Scope, cfg.ScopeID, "log_metrics", meta)
	if err != nil {
		db.Rollback()
		return false, err
	}
	return false, db.Commit().Error
}

func (p *provider) convertToMetricMeta(cfg *db.LogMetricConfig) (*metrics.MetricMeta, error) {
	m := metrics.NewMeta()
	m.Name.Key = cfg.Metric
	m.Name.Name = cfg.Name
	for _, key := range []string{"dice_org_id", "dice_org_name",
		"dice_project_id", "dice_project_name",
		"dice_application_id", "dice_application_name",
		"dice_runtime_id", "dice_runtime_name",
		"dice_service_name", "level"} {
		m.Tags[key] = &metrics.TagDefine{Key: key, Name: key}
	}
	m.Tags["dice_workspace"] = &metrics.TagDefine{
		Key:  "dice_workspace",
		Name: "Workspace",
		Values: []*metrics.ValueDefine{
			{
				Value: "dev",
				Name:  "Develop",
			},
			{
				Value: "test",
				Name:  "Test",
			},
			{
				Value: "staging",
				Name:  "Staging",
			},
			{
				Value: "prod",
				Name:  "Production",
			},
		},
	}
	var ps []*ProcessorConfig
	if err := json.Unmarshal([]byte(cfg.Processors), &ps); err == nil {
		for _, p := range ps {
			byts, _ := json.Marshal(p.Config)
			proc, err := processors.NewProcessor(cfg.Metric, p.Type, byts)
			if err == nil {
				keys := proc.Keys()
				for _, k := range keys {
					if len(k.Name) <= 0 {
						k.Name = k.Key
					}
					m.Fields[k.Key] = k
				}
			}
			break
		}
	}
	return m, nil
}

// DeleteLogMetricConfig .
func (p *provider) DeleteLogMetricConfig(scope, scopeID string, id int64) error {
	db := p.db.Begin()
	c, err := db.LogMetricConfig.QueryByID(scope, scopeID, id)
	if err != nil {
		db.Rollback()
		return err
	}
	err = db.LogMetricConfig.Delete(scope, scopeID, id)
	if err == nil && c != nil {
		err := p.metricq.UnregeistMetricMeta(scope, scopeID, "log_metrics", c.Metric)
		if err != nil {
			db.Rollback()
			return err
		}
	}
	if err != nil {
		db.Rollback()
		return err
	}
	return db.Commit().Error
}
