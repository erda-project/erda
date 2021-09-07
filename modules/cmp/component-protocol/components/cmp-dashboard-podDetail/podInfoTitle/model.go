package podInfoTitle

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type PodInfoTitle struct {
	base.DefaultProvider
	CtxBdl *bundle.Bundle
	SDK    *cptype.SDK
	Props  Props  `json:"props"`
	Type   string `json:"type"`
}

type Props struct {
	Title string `json:"title"`
	Size  string `json:"size"`
}
