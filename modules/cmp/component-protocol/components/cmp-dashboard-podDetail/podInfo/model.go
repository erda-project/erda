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
	Type   string          `json:"type"`
	Data   map[string]Data `json:"data"`
	Props  Props           `json:"props"`
}

type Props struct {
	ColumnNum int     `json:"columnNum"`
	Fields    []Field `json:"fields"`
}

type Data struct {
	Namespace        string `json:"namespace"`
	Survive          string `json:"survive"`
	Ip               string `json:"ip"`
	PodNum           string `json:"podNum"`
	Workload         string `json:"workload"`
	Node             string `json:"node"`
	ContainerRuntime string `json:"containerRuntime"`
	Tag              []Tag  `json:"tag"`
	Desc             []Desc `json:"desc"`
}

type Field struct {
	Label      string               `json:"label"`
	ValueKey   string               `json:"valueKey"`
	RenderType string               `json:"renderType"`
	Operation  map[string]Operation `json:"operation"`
	SpaceNum   int                  `json:"spaceNum"`
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

type Desc struct {
	Label string `json:"label"`
	Group string `json:"group"`
}
