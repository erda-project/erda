package podTitle

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"
)

func (podTitle *PodTitle) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	pod := (*gs)["pod"].(data.Object)
	podTitle.getProps(pod)
	c.Props = podTitle.Props
	return nil
}
func (podTitle *PodTitle) getProps(pod data.Object) {
	podTitle.Props = Props{Title: pod.String("metadata", "name")}
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "podTitle", func() servicehub.Provider {
		return &PodTitle{Type: "Title"}
	})
}
