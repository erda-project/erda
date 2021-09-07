package ContainerTable

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (containerTable *ContainerTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := containerTable.GenComponentState(c); err != nil {
		return err
	}
	bdl := ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	sdk := cputil.SDK(ctx)

	userID := sdk.Identity.UserID
	orgID := sdk.Identity.OrgID

	splits := strings.Split(containerTable.State.PodID, "_")
	if len(splits) != 2 {
		return fmt.Errorf("invalid pod id: %s", containerTable.State.PodID)
	}

	namespace, name := splits[0], splits[1]
	req := &apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        apistructs.K8SPod,
		ClusterName: containerTable.State.ClusterName,
		Name:        name,
		Namespace:   namespace,
	}

	obj, err := bdl.GetSteveResource(req)
	if err != nil {
		return err
	}

	var data []Data
	containerStatuses := obj.Slice("status", "containerStatuses")
	for _, containerStatus := range containerStatuses {
		states := containerStatus.Map("state")
		status := Status{}
		for k := range states {
			status = parseContainerStatus(k)
		}

		data = append(data, Data{
			Status: status,
			Ready:  containerStatus.String("ready"),
			Name:   containerStatus.String("name"),
			Images: Images{
				RenderType: "copyText",
				Value: Value{
					Text: containerStatus.String("image"),
				},
			},
			RestartCount: containerStatus.String("restartCount"),
			Operate: Operate{
				Operations: map[string]Operation{
					"log": {
						Key:    "checkLog",
						Text:   cputil.I18n(ctx, "log"),
						Reload: false,
						Meta: map[string]string{
							"containerName": containerStatus.String("name"),
							"podName":       name,
							"namespace":     namespace,
						},
					},
					"console": {
						Key:    "checkConsole",
						Text:   cputil.I18n(ctx, "console"),
						Reload: false,
						Meta: map[string]string{
							"containerName": containerStatus.String("name"),
							"podName":       name,
							"namespace":     namespace,
						},
					},
				},
				RenderType: "tableOperation",
			},
		})
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].Name < data[j].Name
	})
	containerTable.Data = map[string][]Data{
		"list": data,
	}

	containerTable.Props.Pagination = false
	containerTable.Props.Scroll.X = 1000
	containerTable.Props.Columns = []Column{
		{
			Width:     80,
			DataIndex: "status",
			Title:     cputil.I18n(ctx, "status"),
		},
		{
			Width:     80,
			DataIndex: "ready",
			Title:     cputil.I18n(ctx, "ready"),
		},
		{
			Width:     120,
			DataIndex: "name",
			Title:     cputil.I18n(ctx, "name"),
		},
		{
			Width:     400,
			DataIndex: "images",
			Title:     cputil.I18n(ctx, "images"),
		},
		{
			Width:     80,
			DataIndex: "restartCount",
			Title:     cputil.I18n(ctx, "restartCount"),
		},
		{
			Width:     100,
			DataIndex: "operate",
			Title:     cputil.I18n(ctx, "operate"),
			Fixed:     "right",
		},
	}
	return nil
}

func (containerTable *ContainerTable) GenComponentState(component *cptype.Component) error {
	if component == nil || component.State == nil {
		return nil
	}

	data, err := json.Marshal(component.State)
	if err != nil {
		logrus.Errorf("failed to marshal for eventTable state, %v", err)
		return err
	}
	var state State
	err = json.Unmarshal(data, &state)
	if err != nil {
		logrus.Errorf("failed to unmarshal for eventTable state, %v", err)
		return err
	}
	containerTable.State = state
	return nil
}

func parseContainerStatus(state string) Status {
	status := Status{
		RenderType: "text",
		Value:      "state",
	}
	switch state {
	case "running":
		status.StyleConfig.Color = "green"
	case "waiting":
		status.StyleConfig.Color = "steelblue"
	case "terminated":
		status.StyleConfig.Color = "red"
	}
	return status
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "containerTable", func() servicehub.Provider {
		return &ContainerTable{}
	})
}
