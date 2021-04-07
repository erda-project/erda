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

type NotifyItemRelation struct {
	BaseModel
	NotifyID     int64 `gorm:"index:notify_id"`
	NotifyItemID int64 `gorm:"index:notify_item_id"`
}

// TableName 设置模型对应数据库表名称
func (NotifyItemRelation) TableName() string {
	return "dice_notify_item_relation"
}
