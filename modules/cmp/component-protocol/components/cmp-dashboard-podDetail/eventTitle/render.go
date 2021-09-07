package eventTitle

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (eventTable *EventTitle) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	eventTable.SDK = cputil.SDK(ctx)
	eventTable.Props = Props{
		eventTable.SDK.I18n("relative events"),
	}
	return nil
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "eventTitle", func() servicehub.Provider {
		return &EventTitle{
			Type: "Title",
		}
	})
}
