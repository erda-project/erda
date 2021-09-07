package PodInfo

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type PodInfo struct {
	base.DefaultProvider
	CtxBdl *bundle.Bundle
	SDK    *cptype.SDK

	Type  string          `json:"type"`
	Data  map[string]Data `json:"data,omitempty"`
	Props Props           `json:"props"`
	State State           `json:"state,omitempty"`
}

type Props struct {
	ColumnNum int     `json:"columnNum"`
	Fields    []Field `json:"fields"`
}

type Data struct {
	Namespace   string `json:"namespace"`
	Age         string `json:"age"`
	Ip          string `json:"ip"`
	Workload    string `json:"workload"`
	Node        string `json:"node"`
	Labels      []Tag  `json:"labels"`
	Annotations []Tag  `json:"annotations"`
}

type Field struct {
	Label      string               `json:"label"`
	ValueKey   string               `json:"valueKey"`
	RenderType string               `json:"renderType,omitempty"`
	Operation  map[string]Operation `json:"operation,omitempty"`
	SpaceNum   int                  `json:"spaceNum,omitempty"`
}

type Operation struct {
	Key     string  `json:"key"`
	Reload  bool    `json:"reload"`
	Command Command `json:"command,omitempty"`
}

type Command struct {
	Key     string       `json:"key"`
	Target  string       `json:"target"`
	State   CommandState `json:"state"`
	JumpOut bool         `json:"jump_out"`
}

type CommandState struct {
	Params map[string]string `json:"params"`
}

type Tag struct {
	Label string `json:"label"`
	Group string `json:"group"`
}

type State struct {
	ClusterName string `json:"clusterName,omitempty"`
	PodID       string `json:"podId,omitempty"`
}
