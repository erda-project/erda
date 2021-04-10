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

type GatewayZone struct {
	Name            string `json:"name" xorm:"not null default '' comment('名称') VARCHAR(1024)"`
	Type            string `json:"type" xorm:"not null default '' comment('类型') VARCHAR(16)"`
	KongPolicies    []byte `json:"kong_policies" xorm:"comment('包含的kong策略id') BLOB"`
	IngressPolicies []byte `json:"ingress_policies" xorm:"comment('包含的ingress策略id') BLOB"`
	BindDomain      string `json:"bind_domain" xorm:"comment('绑定的域名') VARCHAR(1024)"`
	DiceOrgId       string `json:"dice_org_id" xorm:"not null comment('dice企业标识id') VARCHAR(32)"`
	DiceProjectId   string `json:"dice_project_id" xorm:"default '' comment('dice项目标识id') VARCHAR(32)"`
	DiceEnv         string `json:"dice_env" xorm:"not null comment('dice应用所属环境') VARCHAR(32)"`
	DiceClusterName string `json:"dice_cluster_name" xorm:"not null comment('dice集群名') VARCHAR(32)"`
	DiceApp         string `json:"dice_app" xorm:"default '' comment('dice应用名') VARCHAR(128)"`
	DiceService     string `json:"dice_service" xorm:"default '' comment('dice服务名') VARCHAR(128)"`
	TenantId        string `json:"tenant_id" xorm:"not null default '' comment('租户id') VARCHAR(128)"`
	PackageApiId    string `json:"package_api_id" xorm:"not null default '' comment('流量入口中指定api的id') VARCHAR(32)"`
	BaseRow         `xorm:"extends"`
}
