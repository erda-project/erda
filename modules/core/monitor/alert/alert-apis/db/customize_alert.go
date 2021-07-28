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

// CustomizeAlertDB .
type CustomizeAlertDB struct {
	*gorm.DB
}

// GetByID .
func (db *CustomizeAlertDB) GetByID(id uint64) (*CustomizeAlert, error) {
	var alert CustomizeAlert
	if err := db.Where("id=?", id).Find(&alert).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &alert, nil
}

// GetByScopeAndScopeIDAndName .
func (db *CustomizeAlertDB) GetByScopeAndScopeIDAndName(scope, scopeID, name string) (*CustomizeAlert, error) {
	var alert CustomizeAlert
	if err := db.
		Where("alert_scope=? AND alert_scope_id=? AND name=?", scope, scopeID, name).
		Find(&alert).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &alert, nil
}

// QueryByScopeAndScopeID .
func (db *CustomizeAlertDB) QueryByScopeAndScopeID(scope, scopeID string, pageNo, pageSize int) ([]*CustomizeAlert, error) {
	var alerts []*CustomizeAlert
	if err := db.
		Where("alert_scope=? AND alert_scope_id=?", scope, scopeID).
		Order("update_time DESC").
		Offset((pageNo - 1) * pageSize).Limit(pageSize).
		Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

// CountByScopeAndScopeID .
func (db *CustomizeAlertDB) CountByScopeAndScopeID(scope, scopeID string) (int, error) {
	var count int
	if err := db.Table(TableCustomizeAlert).
		Where("alert_scope=?", scope).
		Where("alert_scope_id=?", scopeID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Insert .
func (db *CustomizeAlertDB) Insert(alert *CustomizeAlert) error {
	alert.CreateTime = time.Now()
	alert.UpdateTime = time.Now()
	return db.Create(&alert).Error
}

// Update .
func (db *CustomizeAlertDB) Update(alert *CustomizeAlert) error {
	alert.UpdateTime = time.Now()
	return db.Model(&alert).Update(&alert).Error
}

// UpdateEnable .
func (db *CustomizeAlertDB) UpdateEnable(id uint64, enable bool) error {
	return db.Model(&CustomizeAlert{}).
		Where("id=?", id).
		Update("update_time", time.Now()).
		Update("enable", enable).Error
}

// DeleteByID .
func (db *CustomizeAlertDB) DeleteByID(id uint64) error {
	return db.Delete(&CustomizeAlert{ID: id}).Error
}
