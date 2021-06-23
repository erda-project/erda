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

import "github.com/erda-project/erda/apistructs"

// ApplicationPublishItemRelation 应用和发布项关联关系
type ApplicationPublishItemRelation struct {
	BaseModel
	AppID         int64
	PublishItemID int64
	Env           apistructs.DiceWorkspace
	Creator       string
	AK            string
	AI            string
}

// TableName 设置模型对应数据库表名称
func (ApplicationPublishItemRelation) TableName() string {
	return "dice_app_publish_item_relation"
}
