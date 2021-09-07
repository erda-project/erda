package PodDistribution

import (
	"context"
	"fmt"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"
)

func (pd *PodDistribution) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	pods := (*gs)["pods"].([]data.Object)
	pd.SDK = cputil.SDK(ctx)
	d := Data{Total: len(pods)}
	cnt := make(map[string]int)
	for _, pod := range pods {
		cnt[pod.StringSlice("metadata", "fields")[2]]++
	}
	for k, v := range cnt {
		d.Lists = append(d.Lists, pd.ParsePodStatus(k, v))
	}
	c.Data["list"] = d
	return nil
}
func (pd *PodDistribution) ParsePodStatus(state string, cnt int) List {
	status := List{
		Label: pd.SDK.I18n(state) + fmt.Sprintf("%d", cnt),
		Value: cnt,
	}
	switch state {
	case "Completed":
		status.StyleConfig.Color = "steelBlue"
	case "ContainerCreating":
		status.StyleConfig.Color = "orange"
	case "CrashLoopBackOff":
		status.StyleConfig.Color = "red"
	case "Error":
		status.StyleConfig.Color = "maroon"
	case "Evicted":
		status.StyleConfig.Color = "darkgoldenrod"
	case "ImagePullBackOff":
		status.StyleConfig.Color = "darksalmon"
	case "Pending":
		status.StyleConfig.Color = "teal"
	case "Running":
		status.StyleConfig.Color = "lightgreen"
	case "Terminating":
		status.StyleConfig.Color = "brown"
	}
	return status
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "podDistribution", func() servicehub.Provider {
		return &PodDistribution{Type: "LinearDistribution"}
	})
}
