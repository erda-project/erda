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

type GatewayKongInfo struct {
	Az              string `json:"az" xorm:"not null comment('集群名') VARCHAR(32)"`
	Env             string `json:"env" xorm:"default '' comment('环境名') VARCHAR(32)"`
	ProjectId       string `json:"project_id" xorm:"not null comment('项目id') VARCHAR(32)"`
	ProjectName     string `json:"project_name" xorm:"not null comment('项目名') VARCHAR(256)"`
	KongAddr        string `json:"kong_addr" xorm:"not null comment('kong admin地址') VARCHAR(256)"`
	Endpoint        string `json:"endpoint" xorm:"not null comment('kong gateway地址') VARCHAR(256)"`
	NeedUpdate      int    `json:"need_update" xorm:"default 1 comment('待更新标识') TINYINT(1)"`
	InnerAddr       string `json:"inner_addr" xorm:"not null default '' comment('kong内网地址') varchar(1024)"`
	ServiceName     string `josn:"service_name" xorm:"not null default '' comment('kong的服务名称') varchar(32)"`
	AddonInstanceId string `json:"addon_instance_id" xorm:"not null default '' comment('addon id') varchar(32)"`
	TenantId        string `json:"tenant_id" xorm:"not null default '' comment('租户id') VARCHAR(128)"`
	TenantGroup     string `json:"tenant_group" xorm:"not null default '' comment('租户分组') VARCHAR(128)"`
	BaseRow         `xorm:"extends"`
}
