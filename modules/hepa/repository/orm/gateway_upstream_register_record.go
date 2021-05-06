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

type GatewayUpstreamRegisterRecord struct {
	UpstreamId   string `json:"upstream_id" xorm:"not null comment('应用标识id') VARCHAR(32)"`
	RegisterId   string `json:"register_id" xorm:"not null comment('应用注册id') VARCHAR(64)"`
	UpstreamApis []byte `json:"upstream_apis" xorm:"comment('api注册列表') BLOB"`
	BaseRow      `xorm:"extends"`
}
