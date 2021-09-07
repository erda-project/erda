package PodDistribution

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type PodDistribution struct {
	base.DefaultProvider
	CtxBdl *bundle.Bundle
	SDK    *cptype.SDK
	Data   Data   `json:"data"`
	Type   string `json:"type"`
}

type Data struct {
	Total int    `json:"total"`
	Lists []List `json:"list"`
}

type StyleConfig struct {
	Color string `json:"color"`
}
type List struct {
	StyleConfig StyleConfig `json:"styleConfig"`
	Tip         string      `json:"tip"`
	Value       int         `json:"value"`
	Label       string      `json:"label"`
}
