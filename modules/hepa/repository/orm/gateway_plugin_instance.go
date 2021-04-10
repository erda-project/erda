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

type GatewayPluginInstance struct {
	PluginId   string `json:"plugin_id" xorm:"not null default '' comment('插件id') index VARCHAR(128)"`
	PluginName string `json:"plugin_name" xorm:"not null default '' comment('插件名称') VARCHAR(128)"`
	PolicyId   string `json:"policy_id" xorm:"not null default '' comment('策略id') VARCHAR(32)"`
	ConsumerId string `json:"consumer_id" xorm:"comment('消费者id') VARCHAR(32)"`
	GroupId    string `json:"group_id" xorm:"comment('组id') VARCHAR(32)"`
	RouteId    string `json:"route_id" xorm:"comment('路由id') VARCHAR(32)"`
	ServiceId  string `json:"service_id" xorm:"comment('服务id') VARCHAR(32)"`
	ApiId      string `json:"api_id" xorm:"default '' comment('apiID') VARCHAR(32)"`
	BaseRow    `xorm:"extends"`
}
