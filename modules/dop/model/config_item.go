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

package model

import "time"

// ConfigItem 配置信息
type ConfigItem struct {
	ID           int64     `json:"id" gorm:"primary_key"`
	CreatedAt    time.Time `json:"createdAt" gorm:"column:create_time"`
	UpdatedAt    time.Time `json:"updatedAt" gorm:"column:update_time"`
	IsSync       bool      // deprecated
	Dynamic      bool      // deprecated
	Encrypt      bool      // deprecated
	DeleteRemote bool      // deprecated
	IsDeleted    string
	NamespaceID  uint64 `gorm:"index:namespace_id"`
	ItemKey      string
	ItemValue    string
	ItemComment  string
	ItemType     string // FILE, ENV
	Source       string
	Status       string // deprecated
}

// TableName 设置模型对应数据库表名称
func (ConfigItem) TableName() string {
	return "dice_config_item"
}
