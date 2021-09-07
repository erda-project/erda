package ContainerTable

import (
	"context"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/cmp/component-protocol/components/cmp-dashboard-podDetail/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/rancher/wrangler/pkg/data"
)

func (containerTable *ContainerTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	pod := (*gs)["pod"].(data.Object)
	d := make([]Data, 0)
	namespace := pod.String("metadata", "namespace")
	pods := pod.String("metadata", "name")
	for _, container := range pod.Slice("status", "containerStatuses") {
		surviveTime, err := common.SurviveTime(pod.String("state", "running", "startedAt"))
		if err != nil {
			return err
		}
		d = append(d, Data{
			Survive: surviveTime,
			Operate: Operate{RenderType: "tableOperation", Operations: map[string]Operation{"log": {
				Key:     "gotoPod",
				Command: Command{Key: "goto", Target: "log", JumpOut: true},
				Text:    containerTable.SDK.I18n("logPage"),
				Reload:  false,
				State:   CommandState{Params: map[string]string{"namespace": namespace, "pods": pods, "containerName": container.String("name")}},
			}, "console": {
				Key:     "gotoPod",
				Command: Command{Key: "goto", Target: "consolePage", JumpOut: true},
				Text:    containerTable.SDK.I18n("console"),
				Reload:  false,
				State:   CommandState{Params: map[string]string{"namespace": namespace, "pods": pods, "containerName": container.String("name")}},
			}},
			},
			Status: Status{RenderType: "text", Value: pod.String("status", "phase")},
			Ready:  container.String("ready"),
			Name:   container.String("name"),
			Images: Images{
				RenderType: "copyText",
				Value:      Value{Text: container.String("image")},
			},
			RebootTimes: container.String("restartCount"),
		})
	}
	c.Props = containerTable.GetProps()
	c.Data["list"] = d
	return nil
}
func (containerTable *ContainerTable) GetProps() Props {
	p := Props{
		Pagination: false,
		Scroll:     Scroll{X: 1000},
		Columns: []Column{
			{DataIndex: "status", Title: containerTable.SDK.I18n("status"), Width: 120},
			{DataIndex: "ready", Title: containerTable.SDK.I18n("ready"), Width: 120},
			{DataIndex: "name", Title: containerTable.SDK.I18n("name"), Width: 120},
			{DataIndex: "images", Title: containerTable.SDK.I18n("images")},
			{DataIndex: "rebootTimes", Title: containerTable.SDK.I18n("rebootTimes"), Width: 120},
			{DataIndex: "survive", Title: containerTable.SDK.I18n("survive"), Width: 120},
			{DataIndex: "operate", Title: containerTable.SDK.I18n("operate"), Width: 200, Fixed: "right"},
		},
	}
	return p
}
func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "containerTable", func() servicehub.Provider {
		return &ContainerTable{Type: "Table"}
	})
}
