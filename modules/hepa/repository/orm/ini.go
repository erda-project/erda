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

package orm

type Ini struct {
	IniName  string `json:"ini_name" xorm:"not null default '' comment('配置信息名称') index VARCHAR(128)"`
	IniDesc  string `json:"ini_desc" xorm:"not null default '' comment('配置信息介绍') VARCHAR(256)"`
	IniValue string `json:"ini_value" xorm:"not null default '' comment('配置信息参数值') VARCHAR(1024)"`
	BaseRow  `xorm:"extends"`
}
