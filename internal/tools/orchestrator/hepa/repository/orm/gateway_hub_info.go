// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package orm

type GatewayHubInfo struct {
	OrgId           string `json:"org_id" xorm:"not null comment('dice企业标识id') VARCHAR(32)"`
	DiceEnv         string `json:"dice_env" xorm:"not null comment('dice环境') VARCHAR(32)"`
	DiceClusterName string `json:"dice_cluster_name" xorm:"not null comment('dice集群名') VARCHAR(32)"`
	BindDomain      string `json:"bind_domain" xorm:"comment('绑定的域名') VARCHAR(1024)"`
	Description     string `json:"description" xorm:"comment('描述') VARCHAR(256)"`
	BaseRow         `xorm:"extends"`
}
