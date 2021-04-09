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

type GatewayRoute struct {
	RouteId   string `json:"route_id" xorm:"not null default '' comment('路由id') index VARCHAR(128)"`
	Protocols string `json:"protocols" xorm:"comment('协议列表') VARCHAR(128)"`
	Methods   string `json:"methods" xorm:"comment('方法列表') VARCHAR(128)"`
	Hosts     string `json:"hosts" xorm:"comment('主机列表') VARCHAR(256)"`
	Paths     string `json:"paths" xorm:"comment('路径列表') VARCHAR(256)"`
	ServiceId string `json:"service_id" xorm:"not null default '' comment('绑定服务id') VARCHAR(32)"`
	Config    string `json:"config" xorm:"default '' comment('选填配置') VARCHAR(1024)"`
	ApiId     string `json:"api_id" xorm:"not null default '' comment('apiid') VARCHAR(32)"`
	BaseRow   `xorm:"extends"`
}
