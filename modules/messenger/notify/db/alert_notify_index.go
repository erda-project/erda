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
	"time"

	"github.com/jinzhu/gorm"
)

type AlertNotifyIndexDB struct {
	*gorm.DB
}

type AlertNotifyIndex struct {
	ID         int64     `json:"id" gorm:"column:id"`
	NotifyID   int64     `json:"notify_id" gorm:"column:notify_id"`
	NotifyName string    `json:"notify_name" gorm:"column:notify_name"`
	Status     string    `json:"status" gorm:"column:status"`
	Channel    string    `json:"channel" gorm:"column:channel"`
	Attributes string    `json:"attributes" gorm:"column:attributes"`
	ScopeType  string    `json:"scope_type" gorm:"column:scope_type"`
	ScopeID    string    `json:"scope_id" gorm:"column:scope_id"`
	OrgID      int64     `json:"org_id" gorm:"column:org_id"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	SendTime   time.Time `json:"send_time" gorm:"column:send_time"`
}

func (AlertNotifyIndex) TableName() string {
	return "alert_notify_index"
}

func (db *AlertNotifyIndexDB) CreateAlertNotifyIndex(alertNotifyIndex *AlertNotifyIndex) (int64, error) {
	err := db.Create(&alertNotifyIndex).Error
	if err != nil {
		return 0, err
	}
	return alertNotifyIndex.ID, nil
}
