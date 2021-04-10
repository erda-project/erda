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

type GatewayConsumerApi struct {
	ConsumerId string `json:"consumer_id" xorm:"not null default '' comment('消费者id') VARCHAR(32)"`
	ApiId      string `json:"api_id" xorm:"not null default '' comment('apiId') VARCHAR(32)"`
	Policies   string `json:"policies" xorm:"comment('策略信息') VARCHAR(512)"`
	BaseRow    `xorm:"extends"`
}
