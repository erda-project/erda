package Page

import "github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

type Page struct {
	base.DefaultProvider
	Type string `json:"type"`
}
