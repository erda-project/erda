package PodTitle

import (
	"context"
	"fmt"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

func (podTitle *PodTitle) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	podTitle.Type = "Title"
	podTitle.Props.Size = "small"

	total := 0
	for _, count := range podTitle.State.Values {
		total += count
	}
	podTitle.Props.Title = fmt.Sprintf("%s: %d", cputil.I18n(ctx, "podNum"), total)
	return nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "podTitle", func() servicehub.Provider {
		return &PodTitle{}
	})
}
