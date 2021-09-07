package PodDistribution

import (
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type PodDistribution struct {
	base.DefaultProvider

	Data  Data   `json:"data"`
	Type  string `json:"type"`
	State State  `json:"state,omitempty"`
}

type Data struct {
	Total int    `json:"total"`
	Lists []List `json:"list"`
}

type List struct {
	Color string `json:"color"`
	Tip   string `json:"tip"`
	Value int    `json:"value"`
	Label string `json:"label"`
}

type State struct {
	Values map[string]int `json:"values,omitempty"`
}
