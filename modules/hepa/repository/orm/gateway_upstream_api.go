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

type GatewayUpstreamApi struct {
	UpstreamId  string `json:"upstream_id" xorm:"not null comment('应用标识id') VARCHAR(32)"`
	RegisterId  string `json:"register_id" xorm:"not null comment('应用注册id') VARCHAR(64)"`
	ApiName     string `json:"api_name" xorm:"not null comment('标识api的名称，应用下唯一') VARCHAR(256)"`
	Path        string `json:"path" xorm:"not null comment('注册的api路径') VARCHAR(256)"`
	GatewayPath string `json:"gateway_path" xorm:"not null comment('网关api路径') VARCHAR(256)"`
	Method      string `json:"method" xorm:"not null comment('注册的api方法') VARCHAR(256)"`
	Address     string `json:"address" xorm:"not null comment('注册的转发地址') VARCHAR(256)"`
	Doc         []byte `json:"doc" xorm:"comment('api描述') BLOB"`
	ApiId       string `json:"api_id" xorm:"default '' comment('api标识id') VARCHAR(32)"`
	IsInner     int    `json:"is_inner" xorm:"not null default 0 comment('是否是内部api') TINYINT(1)"`
	BaseRow     `xorm:"extends"`
}
