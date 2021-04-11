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

// CustomizeAlertRuleDB .
type CustomizeAlertRuleDB struct {
	*gorm.DB
}

// QueryByAlertIDs .
func (db *CustomizeAlertRuleDB) QueryByAlertIDs(alertIDs []uint64) ([]*CustomizeAlertRule, error) {
	var rules []*CustomizeAlertRule
	if err := db.
		Where("customize_alert_id IN (?)", alertIDs).
		Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// QueryEnabledByScope .
func (db *CustomizeAlertRuleDB) QueryEnabledByScope(scope, scopeID string) ([]*CustomizeAlertRule, error) {
	var rules []*CustomizeAlertRule
	if err := db.
		Where("alert_scope=?", scope).
		Where("alert_scope_id=?", scopeID).
		Where("enable=?", true).
		Find(&rules).Error; err != nil {
		return nil, err
	}

	return rules, nil
}

// QueryEnabledByScopeAndIndices .
func (db *CustomizeAlertRuleDB) QueryEnabledByScopeAndIndices(
	scope, scopeID string, indices []string) ([]*CustomizeAlertRule, error) {
	var rules []*CustomizeAlertRule
	if err := db.
		Where("alert_scope=?", scope).
		Where("alert_scope_id=?", scopeID).
		Where("alert_index IN (?)", indices).
		Where("enable=?", true).
		Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// Insert .
func (db *CustomizeAlertRuleDB) Insert(rule *CustomizeAlertRule) error {
	rule.CreateTime = time.Now()
	rule.UpdateTime = time.Now()
	return db.Create(&rule).Error
}

// Update .
func (db *CustomizeAlertRuleDB) Update(rule *CustomizeAlertRule) error {
	rule.UpdateTime = time.Now()
	return db.Model(&rule).Update(&rule).Error
}

// UpdateEnableByAlertID .
func (db *CustomizeAlertRuleDB) UpdateEnableByAlertID(alertID uint64, enable bool) error {
	return db.Model(&CustomizeAlertRule{}).
		Where("customize_alert_id=?", alertID).
		Update("update_time", time.Now()).
		Update("enable", enable).Error
}

// DeleteByAlertID .
func (db *CustomizeAlertRuleDB) DeleteByAlertID(alertID uint64) error {
	return db.Where("customize_alert_id=?", alertID).Delete(CustomizeAlertRule{}).Error
}

// DeleteByIDs .
func (db *CustomizeAlertRuleDB) DeleteByIDs(ids []uint64) error {
	return db.Where("id IN (?)", ids).Delete(CustomizeAlertRule{}).Error
}
