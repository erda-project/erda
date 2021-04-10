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

type GatewayIngressPolicy struct {
	Name            string `json:"name" xorm:"not null default '' comment('名称') VARCHAR(32)"`
	Regions         string `json:"regions" xorm:"not null default '' comment('作用域') VARCHAR(128)"`
	Az              string `json:"az" xorm:"not null comment('集群名') VARCHAR(32)"`
	ZoneId          string `json:"zone_id" xorm:"default null comment('所属的zone') VARCHAR(32)"`
	Config          []byte `json:"config" xorm:"comment('具体配置') BLOB"`
	ConfigmapOption []byte `json:"configmap_option" xorm:"comment('ingress configmap option') BLOB"`
	MainSnippet     []byte `json:"main_snippet" xorm:"comment('ingress configmap main 配置') BLOB"`
	HttpSnippet     []byte `json:"http_snippet" xorm:"comment('ingress configmap http 配置') BLOB"`
	ServerSnippet   []byte `json:"server_snippet" xorm:"comment('ingress configmap server 配置') BLOB"`
	Annotations     []byte `json:"annotations" xorm:"comment('包含的annotations') BLOB"`
	LocationSnippet []byte `json:"location_snippet" xorm:"comment('nginx location 配置') BLOB"`
	BaseRow         `xorm:"extends"`
}
