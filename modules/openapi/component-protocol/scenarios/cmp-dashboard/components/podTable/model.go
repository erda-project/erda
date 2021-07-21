package podTable

import (
	"github.com/erda-project/erda-infra/base/servicehub"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
)

type PodInfoTable struct {
	CtxBdl     protocol.ContextBundle
	hub        *servicehub.Hub
	Type       string                 `json:"type"`
	Props      map[string]interface{} `json:"props"`
	Operations map[string]interface{} `json:"operations"`
	State      common.State           `json:"state"`
	Data       []RowItem              `json:"data"`
}
type Columns struct {
	Title     string `json:"title"`
	DataIndex string `json:"dataIndex"`
	Width     int    `json:"width,omitempty"`
	Sortable  bool   `json:"sorter"`
}
type Meta struct {
	Id       int `json:"id,omitempty"`
	PageSize int `json:"pageSize,omitempty"`
	PageNo   int `json:"pageNo,omitempty"`
	// todo :
	sortColumn string
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
type Operation struct {
	Key           string      `json:"key"`
	Reload        bool        `json:"reload"`
	FillMeta      string      `json:"fillMeta,omitempty"`
	Target        string      `json:"target,omitempty"`
	Meta          interface{} `json:"meta,omitempty"`
	ClickableKeys interface{} `json:"clickableKeys,omitempty"`
	Command       Command     `json:"command,omitempty"`
}
type Command struct {
	Key     string       `json:"key"`
	Command CommandState `json:"command"`
	Target  string       `json:"target"`
}
type CommandState struct {
	Visible  bool     `json:"visible"`
	FromData FromData `json:"from_data"`
}
type FromData struct {
	RecordId string `json:"record_id"`
}

type labels struct {
	RenderType string               `json:"render_type"`
	Value      []LabelsValue        `json:"value"`
	Operation  map[string]Operation `json:"operation"`
}

type LabelsValue struct {
	Label string `json:"label"`
	Group string `json:"group"`
}
