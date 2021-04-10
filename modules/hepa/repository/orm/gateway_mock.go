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

type GatewayMock struct {
	Az      string `json:"az" xorm:"VARCHAR(256)"`
	HeadKey string `json:"head_key" xorm:"not null pk VARCHAR(256)"`
	Body    string `json:"body" xorm:"TEXT"`
	Pathurl string `json:"pathurl" xorm:"not null pk VARCHAR(256)"`
	Method  string `json:"method" xorm:"VARCHAR(32)"`
	BaseRow `xorm:"extends"`
}

func (row *GatewayMock) GetPK() map[string]interface{} {
	res := row.BaseRow.GetPK()
	res["head_key"] = row.HeadKey
	return res
}
