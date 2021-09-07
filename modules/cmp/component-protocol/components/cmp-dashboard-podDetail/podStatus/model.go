package PodStatus

import (
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type PodStatus struct {
	base.DefaultProvider

	Type  string `json:"type"`
	Props Props  `json:"props"`
	State State  `json:"state,omitempty"`
}

type Props struct {
	Value       string      `json:"value"`
	StyleConfig StyleConfig `json:"styleConfig"`
}

type StyleConfig struct {
	Color string `json:"color"`
}

type State struct {
	ClusterName string `json:"clusterName,omitempty"`
	PodID       string `json:"podId,omitempty"`
}
