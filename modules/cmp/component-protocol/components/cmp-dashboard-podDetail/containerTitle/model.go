package ContainerTitle

import (
	"context"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ContainerTitle struct {
	Ctx    context.Context
	CtxBdl *bundle.Bundle
	base.DefaultProvider
	SDK *cptype.SDK

	Type  string `json:"type"`
	Props Props  `json:"props"`
}

type Props struct {
	Title string `json:"title"`
}
