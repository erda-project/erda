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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/database/dbengine"
)

// Comment 工单评论模型
type Comment struct {
	dbengine.BaseModel

	TicketID    int64
	CommentType apistructs.TCType
	Content     string               `gorm:"type:text"`
	IRComment   apistructs.IRComment `json:"irComment" gorm:"column:ir_comment;type:text"`
	UserID      string
}

// TableName 设置模型对应数据库表名称
func (Comment) TableName() string {
	return "ps_comments"
}
