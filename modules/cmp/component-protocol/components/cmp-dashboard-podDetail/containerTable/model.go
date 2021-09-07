package ContainerTable

import (
	"context"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ContainerTable struct {
	Ctx    context.Context
	CtxBdl *bundle.Bundle
	base.DefaultProvider
	SDK *cptype.SDK

	Type  string            `json:"type"`
	Data  map[string][]Data `json:"data"`
	Props Props             `json:"props"`
}

type Data struct {
	Survive     string  `json:"survive"`
	Operate     Operate `json:"operate"`
	Status      Status  `json:"status"`
	Ready       string  `json:"ready"`
	Name        string  `json:"name"`
	Images      Images  `json:"images"`
	RebootTimes string  `json:"rebootTimes"`
}

type Scroll struct {
	X int `json:"x"`
}

type Props struct {
	Pagination bool     `json:"pagination"`
	Scroll     Scroll   `json:"scroll"`
	Columns    []Column `json:"columns"`
}

type Column struct {
	Width     int    `json:"width"`
	DataIndex string `json:"dataIndex"`
	Title     string `json:"title"`
	Fixed     string `json:"fixed"`
}

type Operate struct {
	Operations map[string]Operation `json:"operations"`
	RenderType string               `json:"renderType"`
}

type Status struct {
	RenderType  string      `json:"renderType"`
	Value       string      `json:"value"`
	StyleConfig StyleConfig `json:"styleConfig"`
}

type Images struct {
	RenderType string `json:"renderType"`
	Value      Value  `json:"value"`
}

type Operation struct {
	Key     string       `json:"key"`
	Command Command      `json:"command"`
	Text    string       `json:"text"`
	Reload  bool         `json:"reload"`
	State   CommandState `json:"state"`
}

type CommandState struct {
	Params map[string]string `json:"params"`
}

type StyleConfig struct {
	Color string `json:"color"`
}

type Value struct {
	Text string `json:"text"`
}

type Command struct {
	JumpOut bool   `json:"jumpOut"`
	Key     string `json:"key"`
	Target  string `json:"target"`
}
