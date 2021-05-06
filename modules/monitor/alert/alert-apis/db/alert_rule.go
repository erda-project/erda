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
	"github.com/jinzhu/gorm"
)

// AlertRuleDB .
type AlertRuleDB struct {
	*gorm.DB
}

// DistinctAlertTypeByScope .
func (db *AlertRuleDB) DistinctAlertTypeByScope(scope string) ([]string, error) {
	rows, err := db.Table(TableAlertRules).
		Where("alert_scope=?", scope).
		Where("enable=?", true).
		Select("distinct(alert_type)").Rows()
	if err != nil {
		return nil, err
	}

	var alertTypes []string
	for rows.Next() {
		var alertType string
		if err := rows.Scan(&alertType); err != nil {
			return nil, err
		}
		alertTypes = append(alertTypes, alertType)
	}
	return alertTypes, nil
}

// QueryEnabledByScope .
func (db *AlertRuleDB) QueryEnabledByScope(scope string) ([]*AlertRule, error) {
	var rules []*AlertRule
	if err := db.
		Where("alert_scope=?", scope).
		Where("enable=?", true).
		Find(&rules).Error; err != nil {
		return nil, err
	}

	return rules, nil
}

// QueryEnabledByScopeAndIndices .
func (db *AlertRuleDB) QueryEnabledByScopeAndIndices(scope string, indices []string) ([]*AlertRule, error) {
	var rules []*AlertRule
	if err := db.
		Where("alert_scope=?", scope).
		Where("alert_index IN (?)", indices).
		Where("enable=?", true).
		Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// QueryByIndexes .
func (db *AlertRuleDB) QueryByIndexes(indexes []string) ([]*AlertRule, error) {
	var rules []*AlertRule
	if err := db.Where("alert_index IN (?)", indexes).Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}
