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

type GatewayGroup struct {
	GroupName   string `json:"group_name" xorm:"not null default '' comment('组名') unique(group_consumer) VARCHAR(128)"`
	DispalyName string `json:"dispaly_name" xorm:"not null default '' comment('展示名称') VARCHAR(128)"`
	ConsumerId  string `json:"consumer_id" xorm:"not null default '' comment('所属消费者id') unique(group_consumer) VARCHAR(32)"`
	Policies    string `json:"policies" xorm:"comment('策略配置，存策略id') VARCHAR(1024)"`
	BaseRow     `xorm:"extends"`
}

func (group GatewayGroup) IsEmpty() bool {
	return len(group.GroupName) == 0
}
