package PodTitle

import (
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type PodTitle struct {
	base.DefaultProvider

	Props Props  `json:"props"`
	Type  string `json:"type"`
	State State  `json:"state,omitempty"`
}

type Props struct {
	Size  string `json:"size"`
	Title string `json:"title"`
}

type State struct {
	Values map[string]int `json:"values"`
}
