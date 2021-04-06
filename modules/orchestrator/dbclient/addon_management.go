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

package dbclient

import (
	"time"
)

// relay on the ops modules（ops/dbclient/addon_management.go）
// remove the dependency when ops complete the migration

// addon management
type AddonManagement struct {
	ID          uint64 `gorm:"primary_key"`
	AddonID     string `gorm:"type:varchar(64)"` // 主键
	Name        string `gorm:"type:varchar(64)"`
	ProjectID   string
	OrgID       string
	AddonConfig string `gorm:"type:text"`
	CPU         float64
	Mem         uint64
	Nodes       int
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
}

func (AddonManagement) TableName() string {
	return "tb_addon_management"
}
