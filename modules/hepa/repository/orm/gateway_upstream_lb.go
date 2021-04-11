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

type GatewayUpstreamLb struct {
	ZoneId           string `json:"zone_id" xorm:"default '' comment('所属的zone') VARCHAR(32)"`
	OrgId            string `json:"org_id" xorm:"not null comment('企业标识id') VARCHAR(32)"`
	ProjectId        string `json:"project_id" xorm:"not null comment('项目标识id') VARCHAR(32)"`
	LbName           string `json:"lb_name" xorm:"not null comment('应用名称') VARCHAR(128)"`
	Env              string `json:"env" xorm:"not null comment('应用所属环境') VARCHAR(32)"`
	Az               string `json:"az" xorm:"not null comment('集群名') VARCHAR(32)"`
	LastDeploymentId int    `json:"last_deployment_id" xorm:"not null comment('最近一次target上线请求的部署id') INT(11)"`
	KongUpstreamId   string `json:"kong_upstream_id" xorm:"comment('kong的upstream_id') VARCHAR(128)"`
	HealthcheckPath  string `json:"healthcheck_path" xorm:"comment('HTTP健康检查路径') VARCHAR(128)"`
	Config           []byte `json:"config" xorm:"comment('负载均衡配置') BLOB"`
	BaseRow          `xorm:"extends"`
}
