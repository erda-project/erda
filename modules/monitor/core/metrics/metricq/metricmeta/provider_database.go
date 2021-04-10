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
	"encoding/json"
	"fmt"
	"time"

	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/providers/i18n"
	"github.com/erda-project/erda/modules/monitor/core/metrics"
	"github.com/jinzhu/gorm"
	"github.com/recallsong/go-utils/reflectx"
)

// tables name
const (
	TableMetricMeta = "sp_metric_meta"
)

type MetricMeta struct {
	ID         int       `gorm:"column:id"`
	Scope      string    `gorm:"column:scope"`
	ScopeID    string    `gorm:"column:scope_id"`
	Group      string    `gorm:"column:group"`
	Metric     string    `gorm:"column:metric"`
	Name       string    `gorm:"column:name"`
	Tags       string    `gorm:"column:tags"`
	Fields     string    `gorm:"column:fields"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
}

func (MetricMeta) TableName() string { return TableMetricMeta }

type DatabaseGroupProvider struct {
	db  *gorm.DB
	log logs.Logger
}

func NewDatabaseGroupProvider(db *gorm.DB, log logs.Logger) (*DatabaseGroupProvider, error) {
	return &DatabaseGroupProvider{
		db:  db,
		log: log,
	}, nil
}

func (p *DatabaseGroupProvider) MappingsByID(id, scope, scopeID string, names []string, ms map[string]*metrics.MetricMeta) (gmm []*GroupMetricMap, err error) {
	for _, name := range names {
		if mm, ok := ms[name]; ok {
			if mm.Labels == nil || mm.Labels["_group"] != "log_metrics" {
				continue
			}
			gmm = append(gmm, &GroupMetricMap{
				Name: mm.Name.Key,
			})
		}
	}
	return gmm, nil
}

func (p *DatabaseGroupProvider) Groups(langCodes i18n.LanguageCodes, t i18n.Translator, scope, scopeID string, ms map[string]*metrics.MetricMeta) (groups []*Group, err error) {
	group := &Group{
		ID:   "log_metrics",
		Name: "Log Metrics",
	}
	for _, m := range ms {
		if m.Labels == nil || m.Labels["_group"] != "log_metrics" {
			continue
		}
		group.Children = append(group.Children, &Group{
			ID:   "log_metrics@" + m.Name.Key,
			Name: m.Name.Name,
		})
	}
	groups = append(groups, group)
	return groups, nil
}

type DatabaseMetaProvider struct {
	db  *gorm.DB
	log logs.Logger
}

func NewDatabaseMetaProvider(db *gorm.DB, log logs.Logger) (*DatabaseMetaProvider, error) {
	return &DatabaseMetaProvider{
		db:  db,
		log: log,
	}, nil
}

func (p *DatabaseMetaProvider) MetricMeta(langCodes i18n.LanguageCodes, i18n i18n.I18n, scope, scopeID string, names ...string) (map[string]*metrics.MetricMeta, error) {
	db := p.db.Table(TableMetricMeta)
	if len(names) <= 0 {
		db = db.Where("`scope`=? AND `scope_id`=?", scope, scopeID)
	} else {
		db = db.Where("`scope`=? AND `scope_id`=? AND `metric` IN (?)", scope, scopeID, names)
	}
	var list []*MetricMeta
	err := db.Find(&list).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]*metrics.MetricMeta, len(list))
	for _, item := range list {
		meta, err := p.convertMetricMetaFromDB(langCodes, i18n, scope, scopeID, item)
		if err != nil {
			p.log.Warn(err)
			continue
		}
		result[item.Metric] = meta
	}
	return result, nil
}

func (p *DatabaseMetaProvider) convertMetricMetaFromDB(langCodes i18n.LanguageCodes, n i18n.I18n, scope, scopeID string, item *MetricMeta) (*metrics.MetricMeta, error) {
	if item == nil || len(item.Metric) <= 0 {
		return nil, fmt.Errorf("invalid metric meta in %s=%s", scope, scopeID)
	}
	meta := &metrics.MetricMeta{
		Name: metrics.NameDefine{
			Key:  item.Metric,
			Name: item.Name,
		},
		Tags:   make(map[string]*metrics.TagDefine),
		Fields: make(map[string]*metrics.FieldDefine),
		Labels: map[string]string{
			"_group": item.Group,
		},
	}
	if len(item.Tags) > 0 {
		err := json.Unmarshal(reflectx.StringToBytes(item.Tags), &meta.Tags)
		if err != nil {
			return nil, fmt.Errorf("invalid tags meta in %s=%s, metric=%s: %s", scope, scopeID, item.Metric, err)
		}
		for k, t := range meta.Tags {
			if t == nil {
				delete(meta.Tags, k)
				continue
			}
			t.Name = n.Text("", langCodes, t.Name)
			for _, v := range t.Values {
				if v == nil {
					continue
				}
				v.Name = n.Text("", langCodes, v.Name)
			}
			t.Key = k
		}
	}
	if len(item.Fields) > 0 {
		err := json.Unmarshal(reflectx.StringToBytes(item.Fields), &meta.Fields)
		if err != nil {
			return nil, fmt.Errorf("invalid fields meta in %s=%s, metric=%s: %s", scope, scopeID, item.Metric, err)
		}
		for k, f := range meta.Fields {
			if f == nil {
				delete(meta.Fields, k)
				continue
			}
			f.Key = k
		}
	}
	return meta, nil
}

var metricMetaRegisterInsertUpdate = "INSERT INTO `" + TableMetricMeta + "`" +
	"(`scope`,`scope_id`,`group`,`metric`,`name`,`tags`,`fields`,`create_time`,`update_time`) " +
	"VALUES(?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE `update_time`=VALUES(`update_time`),`name`=VALUES(`name`),`tags`=VALUES(`tags`),`fields`=VALUES(`fields`)"

func (m *Manager) regeistMetricMeta(scope, scopeID, group string, metrics ...*metrics.MetricMeta) error {
	db := m.db.Begin()
	now := time.Now()
	for _, m := range metrics {
		if m == nil {
			continue
		}
		tags, err := json.Marshal(m.Tags)
		if err != nil {
			db.Rollback()
			return fmt.Errorf("invalid tags: %s", err)
		}
		fields, err := json.Marshal(m.Fields)
		if err != nil {
			db.Rollback()
			return fmt.Errorf("invalid fields: %s", err)
		}
		err = db.Exec(metricMetaRegisterInsertUpdate,
			scope,
			scopeID,
			group,
			m.Name.Key,
			m.Name.Name,
			string(tags),
			string(fields),
			now, now,
		).Error
		if err != nil {
			db.Rollback()
			return err
		}
	}
	return db.Commit().Error
}

func (m *Manager) unregeistMetricMeta(scope, scopeID, group string, metrics ...string) error {
	return m.db.Table(TableMetricMeta).
		Where("`scope`=? AND `scope_id`=? AND `group`=? AND `metric` IN (?)", scope, scopeID, group, metrics).
		Delete(nil).Error
}
