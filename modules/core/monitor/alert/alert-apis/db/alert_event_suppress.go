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

const (
	SuppressTypePause = "pause"
	SuppressTypeStop  = "stop"
)

// AlertEventSuppressDB .
type AlertEventSuppressDB struct {
	*gorm.DB
}

type AlertEventSuppressQueryCondition struct {
	SuppressTypes []string
	EventIds      []string
}

func (db *AlertEventSuppressDB) QueryByCondition(scope, scopeId string, condition *AlertEventSuppressQueryCondition) ([]*AlertEventSuppress, error) {
	query := db.Table(TableAlertEventSuppress).Where("scope=?", scope).Where("scope_id=?", scopeId).Where("enabled=?", true)

	if condition != nil {
		if len(condition.SuppressTypes) > 0 {
			query = query.Where("suppress_type in (?) AND expire_time < now()", condition.SuppressTypes)
		}
		if len(condition.EventIds) > 0 {
			query = query.Where("alert_event_id in (?)", condition.EventIds)
		}
	}

	var list []*AlertEventSuppress
	err := query.Find(&list).Error
	if !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}
	return list, nil
}

func (db *AlertEventSuppressDB) Suppress(orgId int64, scope, scopeId string, eventId string, suppressType string, expireTime time.Time) (bool, error) {
	exists, err := db.QueryByCondition(scope, scopeId, &AlertEventSuppressQueryCondition{
		EventIds: []string{eventId},
	})
	if err != nil {
		return false, err
	}

	var data *AlertEventSuppress
	if len(exists) > 0 {
		data = exists[0]
		data.SuppressType = suppressType
		data.ExpireTime = expireTime
		data.Enabled = true
	} else {
		data = &AlertEventSuppress{
			AlertEventID: eventId,
			OrgID:        orgId,
			Scope:        scope,
			ScopeID:      scopeId,
			SuppressType: suppressType,
			ExpireTime:   expireTime,
			Enabled:      true,
		}
	}

	err = db.Save(data).Error
	return err == nil, err
}

func (db *AlertEventSuppressDB) CancelSuppress(eventId string) (bool, error) {
	query := db.Table(TableAlertEventSuppress).Where("alert_event_id=?", eventId).Update("enabled=?", false)
	if query.Error != nil {
		return false, query.Error
	}
	return true, nil
}
