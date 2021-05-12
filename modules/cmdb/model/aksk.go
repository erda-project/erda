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

	"gorm.io/gorm"
)

type AkSk struct {
	ID          int64          `json:"id" gorm:"primary_key"`
	Ak          string         `json:"ak" gorm:"size:24;unique"`
	Sk          string         `json:"sk" gorm:"size:32;unique"`
	Internal    bool           `json:"internal"`
	Scope       string         `json:"scope"`
	Owner       string         `json:"owner"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"createdAt"`
	DeletedAt   gorm.DeletedAt `json:"deletedAt,omitempty"`
}

func (ak AkSk) TableName() string {
	return "dice_aksks"
}
