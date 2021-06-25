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

import "github.com/erda-project/erda/pkg/database/dbengine"

// Publisher 资源模型
type Publisher struct {
	dbengine.BaseModel

	Name          string // Publisher名称
	PublisherType string // Publisher类型
	PublisherKey  string // PublisherKey，可以作为唯一标示，主要用于监控
	Desc          string // Publisher描述
	Logo          string // Publisher logo地址
	OrgID         int64  // Publisher关联组织ID
	UserID        string `gorm:"column:creator"` // 所属用户Id
}

// TableName 设置模型对应数据库表名称
func (Publisher) TableName() string {
	return "dice_publishers"
}
