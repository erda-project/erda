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

type GatewayZoneInPackage struct {
	PackageId     string `json:"package_id" xorm:"default '' comment('所属的产品包id') VARCHAR(32)"`
	PackageZoneId string `json:"package_zone_id" xorm:"default '' comment('产品包的zone id') VARCHAR(32)"`
	RoutePrefix   string `json:"route_prefix" xorm:"not null comment('路由前缀') VARCHAR(128)"`
	ZoneId        string `json:"zone_id" xorm:"default '' comment('依赖的zone id') VARCHAR(32)"`
	BaseRow       `xorm:"extends"`
}
