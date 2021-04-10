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

type GatewayPolicy struct {
	ZoneId      string `json:"zone_id" xorm:"default '' comment('所属的zone') VARCHAR(32)"`
	PolicyName  string `json:"policy_name" xorm:"default '' comment('策略名称') VARCHAR(128)"`
	DisplayName string `json:"display_name" xorm:"not null default '' comment('策略展示名称') VARCHAR(128)"`
	Category    string `json:"category" xorm:"not null default '' comment('策略类目') VARCHAR(128)"`
	Description string `json:"description" xorm:"not null default '' comment('描述类目') VARCHAR(128)"`
	PluginId    string `json:"plugin_id" xorm:"default '' comment('插件id') VARCHAR(128)"`
	PluginName  string `json:"plugin_name" xorm:"not null default '' comment('插件名称') VARCHAR(128)"`
	Config      []byte `json:"config" xorm:"comment('具体配置') BLOB"`
	ConsumerId  string `json:"consumer_id" xorm:"default null comment('消费者id') VARCHAR(32)"`
	Enabled     int    `json:"enabled" xorm:"default 1 comment('插件开关') TINYINT(1)"`
	ApiId       string `json:"api_id" xorm:"comment('api id') VARCHAR(32)"`
	BaseRow     `xorm:"extends"`
}
