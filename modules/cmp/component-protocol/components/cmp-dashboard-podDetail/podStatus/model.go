package PodStatus

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type PodStatus struct {
	base.DefaultProvider
	CtxBdl *bundle.Bundle
	SDK    *cptype.SDK
	Type   string `json:"type"`
	Props  Props  `json:"props"`
}

type Props struct {
	Value       string      `json:"value"`
	StyleConfig StyleConfig `json:"styleConfig"`
}

type StyleConfig struct {
	Color string `json:"color"`
}
