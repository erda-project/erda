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
	AT_K8S     = "k8s"
	AT_DCOS    = "dcos"
	AT_EDAS    = "edas"
	AT_UNKNOWN = "unknown"
)

type GatewayAzInfo struct {
	Az             string `json:"az" xorm:"not null comment('集群名') VARCHAR(32)"`
	Env            string `json:"env" xorm:"not null comment('应用所属环境') VARCHAR(32)"`
	NeedUpdate     int    `json:"need_update" xorm:"default 1 comment('待更新标识') TINYINT(1)"`
	OrgId          string `json:"org_id" xorm:"not null comment('企业标识id') VARCHAR(32)"`
	ProjectId      string `json:"project_id" xorm:"not null comment('项目标识id') VARCHAR(32)"`
	Type           string `json:"type" xorm:"not null comment('集群类型') VARCHAR(16)"`
	WildcardDomain string `json:"wildcard_domain" xorm:"not null comment('集群泛域名') default '' VARCHAR(1024)"`
	MasterAddr     string `josn:"master_addr" xorm:"not null comment('集群管控地址') default '' VARCHAR(1024)"`
	BaseRow        `xorm:"extends"`
}
