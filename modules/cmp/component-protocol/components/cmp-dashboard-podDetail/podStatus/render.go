package PodStatus

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"
)

func (podStatus *PodStatus) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	podStatus.SDK = cputil.SDK(ctx)
	pod := (*gs)["pod"].(data.Object)
	podStatus.Props = Props{
		StyleConfig: StyleConfig{Color: "green"},
		Value:       pod.StringSlice("metadata", "fields")[2],
	}
	c.Props = podStatus.Props
	return nil
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "podStatus", func() servicehub.Provider {
		return &PodStatus{
			Type: "Text",
		}
	})
}
