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
	POLICY_PROJECT_LEVEL = "project" //已废弃
	POLICY_PACKAGE_LEVEL = "package"
)

type GatewayDefaultPolicy struct {
	Config    []byte `json:"config" xorm:"comment('具体配置') BLOB"`
	DiceApp   string `json:"dice_app" xorm:"default '' comment('dice应用名') VARCHAR(128)"`
	Level     string `json:"level" xorm:"not null comment('策略级别') VARCHAR(32)"`
	Name      string `json:"name" xorm:"not null default '' comment('名称') VARCHAR(32)"`
	TenantId  string `json:"tenant_id" xorm:"not null default '' comment('租户id') VARCHAR(128)"`
	PackageId string `json:"package_id" xorm:"not null default '' comment('流量入口id') VARCHAR(32)"`
	BaseRow   `xorm:"extends"`
}
