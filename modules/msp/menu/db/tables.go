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

const (
	TableTmcIni = "tb_tmc_ini"
)

type TmcIni struct {
	ID         int       `gorm:"column:id;primary_key"`
	IniName    string    `gorm:"column:ini_name"`
	IniDesc    string    `gorm:"column:ini_desc"`
	IniValue   string    `gorm:"column:ini_value"`
	CreateTime time.Time `gorm:"column:create_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
	IsDeleted  string    `gorm:"column:is_deleted"`
}

func (TmcIni) TableName() string { return TableTmcIni }
