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

import "github.com/jinzhu/gorm"

// AlertNotifyTemplateDB .
type AlertNotifyTemplateDB struct {
	*gorm.DB
}

// QueryEnabledByTypesAndIndexes .
func (db *AlertNotifyTemplateDB) QueryEnabledByTypesAndIndexes(
	types, indexes []string) ([]*AlertNotifyTemplate, error) {
	var templates []*AlertNotifyTemplate
	if err := db.
		Where("alert_type IN (?)", types).
		Where("alert_index IN (?)", indexes).
		Where("enable=?", true).
		Find(&templates).Error; err != nil {
		return nil, err
	}
	return templates, nil
}
