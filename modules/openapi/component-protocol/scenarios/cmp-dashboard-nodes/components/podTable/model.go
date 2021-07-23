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

package podTable

import (
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common/table"
)

type PodInfoTable struct {
	table.Table
	Data []RowItem `json:"data"`
}

type RowItem struct {
	ID              string             `json:"id"`
	Status          common.SteveStatus `json:"status"`
	Namespace       string             `json:"namespace"`
	CpuRate         string             `json:"cpu_rate"`
	CpuUsage        string             `json:"cpu_usage"`
	MemRate         string             `json:"mem_rate"`
	MemUsage        string             `json:"mem_usage"`
	RestartTimes    int                `json:"restart_times"`
	ReadyContainers int                `json:"ready_containers"`
	PodIp           string             `json:"ip"`
	Workload        string             `json:"workload"`
	AliveTime       string             `json:"alive_time"`
	NodeName        string             `json:"node_name"`
}
