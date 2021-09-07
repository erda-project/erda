package tableTabs

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

func (tableTabs *TableTabs) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if event.Operation == cptype.InitializeOperation {
		tableTabs.State.ActiveKey = "cpu"
	}
	tableTabs.Props.TabMenu = []TabMenu{
		{
			Key:  "cpu",
			Name: cputil.I18n(ctx, "cpu-analysis"),
		},
		{
			Key:  "mem",
			Name: cputil.I18n(ctx, "mem-analysis"),
		},
	}
	tableTabs.Operations = Operations{
		OnChange: OnChange{
			Key:    "changeTab",
			Reload: true,
		},
	}
	return nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "tableTabs", func() servicehub.Provider {
		return &TableTabs{}
	})
}
