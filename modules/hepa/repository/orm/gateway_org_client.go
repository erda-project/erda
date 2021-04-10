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

type GatewayOrgClient struct {
	OrgId        string `json:"org_id" xorm:"not null VARCHAR(32)"`
	Name         string `json:"name" xorm:"not null default '' comment('消费者名称') VARCHAR(128)"`
	ClientSecret string `json:"client_secret" xorm:"not null default '' comment('客户端凭证') VARCHAR(32)"`
	BaseRow      `xorm:"extends"`
}
