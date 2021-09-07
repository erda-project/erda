package podTitle

import "github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

type PodTitle struct {
	base.DefaultProvider
	Type  string `json:"type"`
	Props Props  `json:"props"`
	State State  `json:"state,omitempty"`
}

type Props struct {
	Title string `json:"title"`
}

type State struct {
	PodID string `json:"podID,omitempty"`
}
