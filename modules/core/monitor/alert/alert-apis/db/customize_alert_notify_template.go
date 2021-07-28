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

// CustomizeAlertNotifyTemplateDB .
type CustomizeAlertNotifyTemplateDB struct {
	*gorm.DB
}

// QueryByAlertIDs .
func (db *CustomizeAlertNotifyTemplateDB) QueryByAlertIDs(alertIDs []uint64) ([]*CustomizeAlertNotifyTemplate, error) {
	var rules []*CustomizeAlertNotifyTemplate
	if err := db.
		Where("customize_alert_id IN (?)", alertIDs).
		Find(&rules).Error; err != nil {
		return nil, err
	}

	return rules, nil
}

// QueryEnabledByTypesAndIndexes .
func (db *CustomizeAlertNotifyTemplateDB) QueryEnabledByTypesAndIndexes(
	types, indexes []string) ([]*CustomizeAlertNotifyTemplate, error) {
	var templates []*CustomizeAlertNotifyTemplate
	if err := db.
		Where("alert_type IN (?)", types).
		Where("alert_index IN (?)", indexes).
		Where("enable=?", true).
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}

// Insert .
func (db *CustomizeAlertNotifyTemplateDB) Insert(notify *CustomizeAlertNotifyTemplate) error {
	notify.CreateTime = time.Now()
	notify.UpdateTime = time.Now()
	return db.Create(&notify).Error
}

// Update .
func (db *CustomizeAlertNotifyTemplateDB) Update(notify *CustomizeAlertNotifyTemplate) error {
	notify.UpdateTime = time.Now()
	return db.Model(&notify).Update(&notify).Error
}

// UpdateEnableByAlertID .
func (db *CustomizeAlertNotifyTemplateDB) UpdateEnableByAlertID(alertID uint64, enable bool) error {
	return db.Model(&CustomizeAlertNotifyTemplate{}).
		Where("customize_alert_id=?", alertID).
		Update("update_time", time.Now()).
		Update("enable", enable).Error
}

// DeleteByAlertID .
func (db *CustomizeAlertNotifyTemplateDB) DeleteByAlertID(alertID uint64) error {
	return db.Where("customize_alert_id=?", alertID).Delete(CustomizeAlertNotifyTemplate{}).Error
}

// DeleteByIDs .
func (db *CustomizeAlertNotifyTemplateDB) DeleteByIDs(ids []uint64) error {
	return db.Where("id IN (?)", ids).Delete(CustomizeAlertNotifyTemplate{}).Error
}
