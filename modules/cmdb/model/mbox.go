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

type MBox struct {
	BaseModel
	Title   string
	Content string
	Label   string //站内信所属模块  monitor|pipeline
	UserID  string
	Status  apistructs.MBoxStatus //read|unread
	OrgID   int64
	ReadAt  *time.Time
}

func (MBox) TableName() string {
	return "dice_mboxs"
}

func (mbox MBox) ToApiData() *apistructs.MBox {
	return &apistructs.MBox{
		ID:        mbox.ID,
		Title:     mbox.Title,
		Content:   mbox.Content,
		Label:     mbox.Label,
		Status:    mbox.Status,
		CreatedAt: mbox.CreatedAt,
		ReadAt:    mbox.ReadAt,
	}
}
