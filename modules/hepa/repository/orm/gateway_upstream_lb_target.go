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

type GatewayUpstreamLbTarget struct {
	LbId         string `json:"lb_id" xorm:"not null comment('关联的lb id') VARCHAR(32)"`
	DeploymentId int    `json:"deployment_id" xorm:"not null comment('上线时的deployment_id') INT(11)"`
	KongTargetId string `json:"kong_target_id" xorm:"comment('kong的target_id') VARCHAR(128)"`
	Target       string `json:"target" xorm:"comment('目的地址') VARCHAR(64)"`
	Healthy      int    `json:"healthy" xorm:"not null default 1 comment('是否健康') TINYINT(1)"`
	Weight       int    `json:"weight" xorm:"not null default 100 comment('轮询权重') INT(11)"`
	BaseRow      `xorm:"extends"`
}
