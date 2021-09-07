package Page

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (page *Page) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	return nil
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "page", func() servicehub.Provider {
		return &Page{Type: "Container"}
	})
}
