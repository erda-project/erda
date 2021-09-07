package podInfoTitle

import (
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type PodInfoTitle struct {
	base.DefaultProvider

	Props Props  `json:"props"`
	Type  string `json:"type"`
}

type Props struct {
	Title string `json:"title"`
	Size  string `json:"size"`
}
