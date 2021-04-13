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

// AlertRecordDB .
type AlertRecordDB struct {
	*gorm.DB
}

// GetByGroupID .
func (db *AlertRecordDB) GetByGroupID(groupID string) (*AlertRecord, error) {
	var record AlertRecord
	if err := db.Where("group_id=?", groupID).Find(&record).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return &record, nil
}

// QueryByCondition .
func (db *AlertRecordDB) QueryByCondition(scope, scopeKey string,
	alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string, pageNo, pageSize uint) (
	[]*AlertRecord, error) {
	var records []*AlertRecord
	s := db.Where("scope = ?", scope).Where("scope_key = ?", scopeKey)
	if len(alertGroups) > 0 {
		where := "alert_group in (?)" + " or ("
		values := []interface{}{alertGroups}
		for i, group := range alertGroups {
			if i > 0 {
				where += " or "
			}
			where += "alert_group like ?"
			values = append(values, group+"%")
		}
		where += ")"
		s = s.Where(where, values...)
	}
	if len(alertStates) > 0 {
		s = s.Where("alert_state in (?)", alertStates)
	}
	if len(alertTypes) > 0 {
		s = s.Where("alert_type in (?)", alertTypes)
	}
	if len(handlerIDs) > 0 {
		s = s.Where("handler_id in (?)", handlerIDs)
	}
	if len(handleStates) > 0 {
		s = s.Where("handle_state in (?)", handleStates)
	}
	if err := s.
		Order("update_time DESC").
		Offset((pageNo - 1) * pageSize).Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

// CountByCondition .
func (db *AlertRecordDB) CountByCondition(scope, scopeKey string,
	alertGroups, alertStates, alertTypes, handleStates, handlerIDs []string) (int, error) {
	var count int
	s := db.Table(TableAlertRecord).
		Where("scope = ?", scope).
		Where("scope_key = ?", scopeKey)
	if len(alertGroups) > 0 {
		where := "alert_group in (?)" + " or ("
		values := []interface{}{alertGroups}
		for i, group := range alertGroups {
			if i > 0 {
				where += " or "
			}
			where += "alert_group like ?"
			values = append(values, group+"%")
		}
		where += ")"
		s = s.Where(where, values...)
	}
	if len(alertStates) > 0 {
		s = s.Where("alert_state in (?)", alertStates)
	}
	if len(alertTypes) > 0 {
		s = s.Where("alert_type in (?)", alertTypes)
	}
	if len(handlerIDs) > 0 {
		s = s.Where("handler_id in (?)", handlerIDs)
	}
	if len(handleStates) > 0 {
		s = s.Where("handle_state in (?)", handleStates)
	}
	if err := s.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// UpdateHandle .
func (db *AlertRecordDB) UpdateHandle(groupID string, issueID uint64, handlerID string, handleState string) error {
	s := db.Table(TableAlertRecord).
		Where("group_id=?", groupID).
		Update("issue_id", issueID).
		Update("handle_time", time.Now())
	if handlerID != "" {
		s.Update("handler_id", handlerID)
	}
	if handleState != "" {
		s.Update("handle_state", handleState)
	}
	return s.Error
}
