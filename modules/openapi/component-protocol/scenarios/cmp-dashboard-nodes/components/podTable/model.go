package podTable

import (
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
)

type PodInfoTable struct {
	common.Table
	Data       []RowItem              `json:"data"`
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
