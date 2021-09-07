package tableTabs

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

func (tableTabs *TableTabs) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	return nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "tableTabs", func() servicehub.Provider {
		return &TableTabs{}
	})
}
