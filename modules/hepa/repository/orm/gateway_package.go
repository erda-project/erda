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

const (
	OPENAPI_SCENE = "openapi"
	WEBAPI_SCENE  = "webapi"
	UNITY_SCENE   = "unity"
)

type GatewayPackage struct {
	DiceOrgId          string `json:"dice_org_id" xorm:"not null comment('dice企业标识id') VARCHAR(32)"`
	DiceProjectId      string `json:"dice_project_id" xorm:"default '' comment('dice项目标识id') VARCHAR(32)"`
	DiceEnv            string `json:"dice_env" xorm:"not null comment('dice环境') VARCHAR(32)"`
	DiceClusterName    string `json:"dice_cluster_name" xorm:"not null comment('dice集群名') VARCHAR(32)"`
	ZoneId             string `json:"zone_id" xorm:"comment('所属的zone') VARCHAR(32)"`
	PackageName        string `json:"package_name" xorm:"not null comment('产品包名称') VARCHAR(32)"`
	BindDomain         string `json:"bind_domain" xorm:"comment('绑定的域名') VARCHAR(1024)"`
	Description        string `json:"description" xorm:"comment('描述') VARCHAR(256)"`
	AclType            string `json:"acl_type" xorm:"not null default 'off' comment('授权方式') VARCHAR(16)"`
	AuthType           string `json:"auth_type" xorm:"not null default '' comment('鉴权方式') VARCHAR(16)"`
	Scene              string `json:"scene" xorm:"not null default '' comment('场景') VARCHAR(32)"`
	RuntimeServiceId   string `json:"runtime_service_id" xorm:"not null default '' comment('关联的service的id') VARCHAR(32)"`
	CloudapiInstanceId string `json:"cloudapi_instance_id" xorm:"not null default '' comment('阿里云API网关的实例id') VARCHAR(128)"`
	CloudapiGroupId    string `json:"cloudapi_group_id" xorm:"not null default '' comment('阿里云API网关的分组id') VARCHAR(128)"`
	CloudapiVpcGrant   string `json:"cloudapi_vpc_grant" xorm:"not null default '' comment('阿里云API网关的VPC Grant') VARCHAR(128)"`
	CloudapiDomain     string `json:"cloudapi_domain" xorm:"not null default '' comment('阿里云API网关的分组二级域名') VARCHAR(1024)"`
	CloudapiNeedBind   int    `json:"cloudapi_need_bind" xorm:"default 0 comment('是否需要绑定阿里云API网关') TINYINT(1)"`
	BaseRow            `xorm:"extends"`
}
