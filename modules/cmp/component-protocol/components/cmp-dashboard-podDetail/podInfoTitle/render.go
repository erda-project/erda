package podInfoTitle

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (podInfoTitle *PodInfoTitle) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	podInfoTitle.Props.Title = cputil.I18n(ctx, "podInfoTitle")
	podInfoTitle.Props.Size = "small"
	return nil
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "podInfoTitle", func() servicehub.Provider {
		return &PodInfoTitle{}
	})
}
