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

type GatewayApi struct {
	ZoneId           string `json:"zone_id" xorm:"default '' comment('所属的zone') VARCHAR(32)"`
	ConsumerId       string `json:"consumer_id" xorm:"not null default '' comment('消费者id') VARCHAR(32)"`
	ApiPath          string `json:"api_path" xorm:"not null default '' comment('api路径') VARCHAR(256)"`
	Method           string `json:"method" xorm:"not null default '' comment('方法') VARCHAR(128)"`
	RedirectAddr     string `json:"redirect_addr" xorm:"not null default '' comment('转发地址') VARCHAR(256)"`
	Description      string `json:"description" xorm:"comment('描述') VARCHAR(256)"`
	GroupId          string `json:"group_id" xorm:"not null default '' comment('服务Id') VARCHAR(32)"`
	Policies         string `json:"policies" xorm:"comment('策略配置') VARCHAR(1024)"`
	UpstreamApiId    string `json:"upstream_api_id" xorm:"not null default '' comment('对应的后端api') VARCHAR(32)"`
	DiceApp          string `json:"dice_app" xorm:"default '' comment('dice应用名') varchar(128)"`
	DiceService      string `json:"dice_service" xorm:"default '' comment('dice服务名') varchar(128)"`
	RegisterType     string `json:"register_type" xorm:"not null default 'auto' comment('注册类型') varchar(16)"`
	NetType          string `json:"net_type" xorm:"not null default 'outer' comment('网络类型') varchar(16)"`
	NeedAuth         int    `json:"need_auth" xorm:"not null default 0 comment('需要鉴权标识') TINYINT(1)"`
	RedirectType     string `json:"redirect_type" xorm:"not null default 'url' comment('转发类型') VARCHAR(32)"`
	RuntimeServiceId string `json:"runtime_service_id" xorm:"not null default '' comment('关联的service的id') VARCHAR(32)"`
	Swagger          []byte `json:"swagger" xorm:"comment('swagger文档') BLOB"`
	BaseRow          `xorm:"extends"`
}
