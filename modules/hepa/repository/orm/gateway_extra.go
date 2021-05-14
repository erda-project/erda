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

type GatewayExtra struct {
	Field   string `json:"field" xorm:"not null default '' comment('域') index(key) VARCHAR(128)"`
	KeyId   string `json:"key_id" xorm:"not null default '' comment('键') index(key) VARCHAR(64)"`
	Value   string `json:"value" xorm:"not null default '' comment('值') VARCHAR(128)"`
	BaseRow `xorm:"extends"`
}
