package PodTitle

import (
	"context"
	"fmt"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

func (podTitle *PodTitle) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	pods := (*gs)["pods"].([]data.Object)
	podTitle.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	podTitle.SDK = cputil.SDK(ctx)
	podTitle.Type = "Title"
	podTitle.Props.Size = "small"
	podTitle.Props.Title = podTitle.SDK.I18n("podNum") + fmt.Sprintf("%d", len(pods))
	return nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "podTitle", func() servicehub.Provider {
		return &PodTitle{}
	})
}
