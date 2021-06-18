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

import (
	"time"

	"github.com/erda-project/erda/apistructs"
)

// ErrorLog 错误日志模型
type ErrorLog struct {
	BaseModel
	ResourceType   apistructs.ErrorResourceType `gorm:"column:resource_type"`
	ResourceID     string                       `gorm:"column:resource_id"`
	Level          apistructs.ErrorLogLevel     `gorm:"column:level"`
	OccurrenceTime time.Time                    `gorm:"column:occurrence_time"`
	HumanLog       string                       `gorm:"column:human_log"`
	PrimevalLog    string                       `gorm:"column:primeval_log"`
	DedupID        string                       `gorm:"column:dedup_id"`
}

// TableName 设置模型对应数据库表名称
func (ErrorLog) TableName() string {
	return "dice_error_box"
}
