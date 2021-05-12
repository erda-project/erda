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

import "time"

type NotifyRecord struct {
	NotifyId    string    `json:"notify_id" gorm:"column:notify_id"`
	NotifyName  string    `json:"notify_name" gorm:"column:notify_name"`
	ScopeType   string    `json:"scope_type" gorm:"column:scope_type"`
	ScopeId     int64     `json:"scope_id" gorm:"column:scope_id"`
	GroupId     string    `json:"group_id" gorm:"column:group_id"`
	NotifyGroup string    `json:"notify_group" gorm:"column:notify_group"`
	Title       string    `json:"title" gorm:"column:title"`
	NotifyTime  time.Time `json:"notify_time" gorm:"column:notify_time"`
	CreateTime  time.Time `json:"create_time" gorm:"column:create_time"`
	UpdateTime  time.Time `json:"update_time" gorm:"column:update_time"`
}

func (n *NotifyRecord) TableName() string {
	return "sp_notify_record"
}
