package PodDistribution

import (
	"context"
	"fmt"
	"sort"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (pd *PodDistribution) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	total := 0
	pd.Data.Lists = nil
	for state, count := range pd.State.Values {
		total += count
		pd.Data.Lists = append(pd.Data.Lists, pd.ParsePodStatus(ctx, state, count))
	}
	pd.Data.Total = total
	sort.Slice(pd.Data.Lists, func(i, j int) bool {
		return pd.Data.Lists[i].Label < pd.Data.Lists[j].Label
	})
	return nil
}

func (pd *PodDistribution) ParsePodStatus(ctx context.Context, state string, cnt int) List {
	status := List{
		Tip:   fmt.Sprintf("%s %d/%d", cputil.I18n(ctx, state), cnt, pd.Data.Total),
		Value: cnt,
		Label: fmt.Sprintf("%s %d", cputil.I18n(ctx, state), cnt),
	}
	switch state {
	case "Completed":
		status.Color = "steelBlue"
	case "ContainerCreating":
		status.Color = "orange"
	case "CrashLoopBackOff":
		status.Color = "red"
	case "Error":
		status.Color = "maroon"
	case "Evicted":
		status.Color = "darkgoldenrod"
	case "ImagePullBackOff":
		status.Color = "darksalmon"
	case "Pending":
		status.Color = "teal"
	case "Running":
		status.Color = "lightgreen"
	case "Terminating":
		status.Color = "brown"
	}
	return status
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-pods", "podDistribution", func() servicehub.Provider {
		return &PodDistribution{Type: "LinearDistribution"}
	})
}
