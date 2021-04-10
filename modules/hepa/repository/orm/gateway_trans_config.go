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

type GatewayTransConfig struct {
	EnvType       string `json:"env_type" xorm:"VARCHAR(32)"`
	RuntimeId     string `json:"runtime_id" xorm:"VARCHAR(64)"`
	ClusterName   string `json:"cluster_name" xorm:"VARCHAR(32)"`
	TargetKey     string `json:"target_key" xorm:"VARCHAR(512)"`
	TargetType    string `json:"target_type" xorm:"VARCHAR(64)"`
	DubboTarget   string `json:"dubbo_target" xorm:"default '' VARCHAR(4096)"`
	OperationType string `json:"operation_type" xorm:"VARCHAR(64)"`
	ZkUrl         string `json:"zk_url" xorm:"VARCHAR(256)"`
	BaseRow       `xorm:"extends"`
}
