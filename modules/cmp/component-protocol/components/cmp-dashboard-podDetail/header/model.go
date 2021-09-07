package Header

import (
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type Header struct {
	base.DefaultProvider
	SDK *cptype.SDK
	CtxBdl *bundle.Bundle
	Type   string `json:"type"`
}
