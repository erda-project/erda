package ContainerTitle

import "github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

type ContainerTitle struct {
	base.DefaultProvider

	Type  string `json:"type"`
	Props Props  `json:"props"`
}

type Props struct {
	Title string `json:"title"`
}
