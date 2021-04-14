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

// AlertDB .
type AlertDB struct {
	*gorm.DB
}

// GetByID .
func (db *AlertDB) GetByID(id uint64) (*Alert, error) {
	var alert Alert
	if err := db.Where("id=?", id).Find(&alert).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &alert, nil
}

// GetByScopeAndScopeIDAndName .
func (db *AlertDB) GetByScopeAndScopeIDAndName(scope, scopeID, name string) (*Alert, error) {
	var alert Alert
	if err := db.
		Where("alert_scope=?", scope).
		Where("alert_scope_id=?", scopeID).
		Where("name=?", name).
		Find(&alert).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &alert, nil
}

// QueryByScopeAndScopeID .
func (db *AlertDB) QueryByScopeAndScopeID(scope, scopeID string, pageNo, pageSize uint64) ([]*Alert, error) {
	var alerts []*Alert
	if err := db.
		Where("alert_scope=?", scope).
		Where("alert_scope_id=?", scopeID).
		Order("id DESC").
		Offset((pageNo - 1) * pageSize).Limit(pageSize).
		Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// CountByScopeAndScopeID .
func (db *AlertDB) CountByScopeAndScopeID(scope, scopeID string) (int, error) {
	var count int
	if err := db.
		Table(TableAlert).
		Where("alert_scope=?", scope).
		Where("alert_scope_id=?", scopeID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Insert .
func (db *AlertDB) Insert(alert *Alert) error {
	alert.Created = time.Now()
	alert.Updated = time.Now()
	return db.Create(&alert).Error
}

// Update .
func (db *AlertDB) Update(alert *Alert) error {
	alert.Updated = time.Now()
	return db.Model(&alert).Update(&alert).Error
}

// UpdateEnable .
func (db *AlertDB) UpdateEnable(id uint64, enable bool) error {
	return db.Table(TableAlert).
		Where("id=?", id).
		Update("updated", time.Now()).
		Update("enable", enable).Error
}

// DeleteByID .
func (db *AlertDB) DeleteByID(id uint64) error {
	return db.Delete(&Alert{ID: id}).Error
}
