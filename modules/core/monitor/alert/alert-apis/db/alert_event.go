// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"github.com/jinzhu/gorm"
)

// AlertEventDB .
type AlertEventDB struct {
	*gorm.DB
}

func (db *AlertEventDB) CreateAlertEvent(data *AlertEvent) error {
	return db.Create(data).Error
}

func (db *AlertEventDB) UpdateAlertEvent(id string, fields map[string]interface{}) error {
	return db.Table(TableAlertEvent).Updates(fields).Where("id=?", id).Error
}

func (db *AlertEventDB) GetById(id string) (*AlertEvent, error) {
	var record AlertEvent
	err := db.Where("id=?", id).Find(&record).Error
	if err == nil {
		return &record, nil
	}
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return nil, err
}

// GetByAlertGroupID .
func (db *AlertEventDB) GetByAlertGroupID(groupID string) (*AlertEvent, error) {
	var record AlertEvent
	if err := db.Where("alert_group_id=?", groupID).Find(&record).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// QueryByCondition .
func (db *AlertEventDB) QueryByCondition(scope, scopeKey string) {
	panic("implement me")
}

// CountByCondition .
func (db *AlertEventDB) CountByCondition(scope, scopeKey string) {
	panic("implement me")
}
