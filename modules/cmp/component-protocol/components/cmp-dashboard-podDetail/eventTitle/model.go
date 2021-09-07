package eventTitle

import (
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

type EventTitle struct {
	base.DefaultProvider

	SDK   *cptype.SDK
	Type  string `json:"type"`
	Props Props  `json:"props"`
}

type Props struct {
	Title string `json:"title"`
}
