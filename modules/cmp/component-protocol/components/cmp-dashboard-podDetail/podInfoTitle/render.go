package podInfoTitle

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-podDetail/common"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (podInfoTitle *PodInfoTitle) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	podInfoTitle.CtxBdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	podInfoTitle.SDK = cputil.SDK(ctx)
	err := common.Transfer(*c, &podInfoTitle.Props)
	if err != nil {
		return err
	}
	podInfoTitle.Props.Title = podInfoTitle.SDK.I18n("podInfo")
	c.Props = podInfoTitle.Props
	return nil
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "podInfoTitle", func() servicehub.Provider {
		return &PodInfoTitle{
			Props: Props{
				Size: "small",
			},
			Type: "Title",
		}
	})
}
