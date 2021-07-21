package cpuTable

import (
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CpuInfoTable struct {
	CtxBdl     protocol.ContextBundle
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
	SortColumn string `json:"sorter"`
}

type RowItem struct {
	ID      string             `json:"id"`
	Status  common.SteveStatus `json:"status"`
	Node    Node               `json:"node"`
	Role    string             `json:"role"`
	Version string             `json:"version"`
	//
	Distribution     Distribution         `json:"distribution"`
	Usage            Distribution         `json:"use"`
	DistributionRate Distribution         `json:"distribution_rate"`
	Labels           Labels               `json:"Labels"`
	Operations       map[string]Operation `json:"operations"`
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
	RenderType string
	Value      string
	Operation  map[string]Operation
	reload     bool
}
type Labels struct {
	RenderType string               `json:"render_type"`
	Value      []LabelsValue        `json:"value"`
	Operation  map[string]Operation `json:"operation"`
}
type Distribution struct {
	RenderType string            `json:"render_type"`
	Value      DistributionValue `json:"value"`
	Status     common.UsageStatusEnum
}

type DistributionValue struct {
	Text    string `json:"text"`
	Percent int    `json:"percent"`
}
type LabelsValue struct {
	Label string `json:"label"`
	Group string `json:"group"`
}
type SteveData struct {
	metav1.TypeMeta
	Metadata metav1.ObjectMeta `json:"metadata,omitempty"`
	Status   SteveDataStatus   `json:"status,omitempty"`
	Spec     SteveDataSpec     `json:"spec,omitempty"`
}
type SteveDataStatus struct {
	Type string `json:"type"`
	Allocatable map[string]interface{} `json:"allocatable"`
	Capacity map[string]interface{} `json:"capacity"`
	Conditions [] metav1.Condition `json:"conditions"`
}
type SteveDataSpec struct {
	Type string `json:"type"`
}
