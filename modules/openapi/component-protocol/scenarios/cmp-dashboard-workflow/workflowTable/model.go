package workflowTable

import (
	"github.com/erda-project/erda-infra/base/servicehub"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-nodes/common"
)

type WorkflowTable struct {
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
	Status  common.SteveStatus `json:"status"`
	Node    Node               `json:"node"`
	Role    string             `json:"role"`
	Version string             `json:"version"`
	// memory from pods divide node allocate memory
	Distribution     Distribution `json:"distribution"`
	Usage            Distribution `json:"use"`
	DistributionRate Distribution `json:"distribution_rate"`
	Labels           labels       `json:"labels"`
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

type Node struct {
	RenderType string               `json:"render_type"`
	Value      string               `json:"value"`
	Operation  map[string]Operation `json:"operation"`
	Reload     bool                 `json:"reload"`
}
type labels struct {
	RenderType string               `json:"render_type"`
	Value      []LabelsValue        `json:"value"`
	Operation  map[string]Operation `json:"operation"`
}
type Distribution struct {
	RenderType string                 `json:"render_type"`
	Value      DistributionValue      `json:"value"`
	Status     common.UsageStatusEnum `json:"status"`
}
type DistributionValue struct {
	Text    string `json:"text"`
	Percent int    `json:"percent"`
}
type LabelsValue struct {
	Label string `json:"label"`
	Group string `json:"group"`
}
