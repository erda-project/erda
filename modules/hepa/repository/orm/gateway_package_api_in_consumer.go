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

type GatewayPackageApiInConsumer struct {
	ConsumerId   string `json:"consumer_id" xorm:"not null default '' comment('消费者id') VARCHAR(32)"`
	PackageId    string `json:"package_id" xorm:"not null default '' comment('产品包id') VARCHAR(32)"`
	PackageApiId string `json:"package_api_id" xorm:"not null default '' comment('产品包 apiid') VARCHAR(32)"`
	BaseRow      `xorm:"extends"`
}
