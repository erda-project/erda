package ContainerTitle

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (containerTitle *ContainerTitle) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	containerTitle.Props.Title = cputil.I18n(ctx, "container")
	return nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "containerTitle", func() servicehub.Provider {
		return &ContainerTitle{}
	})
}
