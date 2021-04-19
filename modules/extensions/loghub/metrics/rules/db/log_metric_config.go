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

package db

import (
	"time"

	"github.com/jinzhu/gorm"
)

// LogMetricConfigTable .
const LogMetricConfigTable = "sp_log_metric_config"

// LogMetricConfig .
type LogMetricConfig struct {
	ID         int64     `gorm:"column:id" json:"id"`
	OrgID      int64     `gorm:"column:org_id" json:"org_id"`
	Scope      string    `gorm:"column:scope" json:"scope"`
	ScopeID    string    `gorm:"column:scope_id" json:"scope_id"`
	Name       string    `gorm:"column:name" json:"name"`
	Metric     string    `gorm:"column:metric" json:"metric"`
	Filters    string    `gorm:"column:filters" json:"filters"`
	Processors string    `gorm:"column:processors" json:"processors"`
	Enable     bool      `gorm:"column:enable" json:"enable"`
	CreateTime time.Time `gorm:"column:create_time" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time" json:"update_time"`
}

// TableName .
func (LogMetricConfig) TableName() string { return LogMetricConfigTable }

// LogMetricConfigDB .
type LogMetricConfigDB struct {
	*gorm.DB
}

// QueryByScope .
func (db *LogMetricConfigDB) QueryByScope(scope, scopeID string) ([]*LogMetricConfig, error) {
	var list []*LogMetricConfig
	if err := db.Table(LogMetricConfigTable).
		Select("`id`,`org_id`,`name`,`metric`,`processors`,`enable`,`create_time`,`update_time`").
		Where("`scope`=? AND `scope_id`=?", scope, scopeID).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// QueryByID .
func (db *LogMetricConfigDB) QueryByID(scope, scopeID string, id int64) (*LogMetricConfig, error) {
	var c LogMetricConfig
	if err := db.Table(LogMetricConfigTable).
		Where("`scope`=? AND `scope_id`=? AND `id`=?", scope, scopeID, id).
		Find(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// Insert .
func (db *LogMetricConfigDB) Insert(cfg *LogMetricConfig) error {
	now := time.Now()
	cfg.ID = 0
	cfg.Enable = true
	cfg.CreateTime = now
	cfg.UpdateTime = now
	return db.Table(LogMetricConfigTable).Create(cfg).Error
}

// Enable .
func (db *LogMetricConfigDB) Enable(scope, scopeID string, id int64, enable bool) error {
	return db.Table(LogMetricConfigTable).
		Where("`scope`=? AND `scope_id`=? AND `id`=?", scope, scopeID, id).
		Updates(map[string]interface{}{
			"enable": enable,
		}).Error
}

// Update .
func (db *LogMetricConfigDB) Update(cfg *LogMetricConfig) error {
	cfg.UpdateTime = time.Now()
	return db.Table(LogMetricConfigTable).
		Where("`scope`=? AND `scope_id`=? AND `id`=?", cfg.Scope, cfg.ScopeID, cfg.ID).
		Updates(map[string]interface{}{
			"name":        cfg.Name,
			"filters":     cfg.Filters,
			"processors":  cfg.Processors,
			"update_time": cfg.UpdateTime,
		}).Error
}

// Delete .
func (db *LogMetricConfigDB) Delete(scope, scopeID string, id int64) error {
	return db.Table(LogMetricConfigTable).
		Where("`scope`=? AND `scope_id`=? AND `id`=?", scope, scopeID, id).
		Delete(nil).Error
}

// QueryEnabledByScope .
func (db *LogMetricConfigDB) QueryEnabledByScope(scope, scopeID string) (list []*LogMetricConfig, err error) {
	if len(scopeID) <= 0 {
		err = db.Table(LogMetricConfigTable).
			Where("`enable`=1 AND `scope`=?", scope).
			Find(&list).Error
	} else {
		err = db.Table(LogMetricConfigTable).
			Where("`enable`=1 AND `scope`=? AND `scope_id`=?", scope, scopeID).
			Find(&list).Error
	}
	return list, err
}
