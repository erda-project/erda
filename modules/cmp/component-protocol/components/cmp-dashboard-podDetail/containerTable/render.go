// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ContainerTable

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmp/cmp_interface"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

var steveServer cmp_interface.SteveServer

func (containerTable *ContainerTable) Init(ctx servicehub.Context) error {
	server, ok := ctx.Service("cmp").(cmp_interface.SteveServer)
	if !ok {
		return errors.New("failed to init component, cmp service in ctx is not a steveServer")
	}
	steveServer = server
	return containerTable.DefaultProvider.Init(ctx)
}

func (containerTable *ContainerTable) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := containerTable.GenComponentState(c); err != nil {
		return err
	}
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

	resp, err := steveServer.GetSteveResource(ctx, req)
	if err != nil {
		return err
	}
	obj := resp.Data()

	var data []Data
	containerStatuses := obj.Slice("status", "containerStatuses")
	for _, containerStatus := range containerStatuses {
		states := containerStatus.Map("state")
		status := Status{}
		for k := range states {
			status = parseContainerStatus(ctx, k)
		}

		containerId := strings.TrimPrefix(containerStatus.String("containerID"), "docker://")
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
							"containerId":   containerId,
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

	containerTable.Props.SortDirections = []string{"descend", "ascend"}
	containerTable.Props.IsLoadMore = true
	containerTable.Props.RowKey = "name"
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
	containerTable.Transfer(c)
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

func (containerTable *ContainerTable) Transfer(component *cptype.Component) {
	component.Props = containerTable.Props
	component.Data = map[string]interface{}{}
	for k, v := range containerTable.Data {
		component.Data[k] = v
	}
	component.State = map[string]interface{}{
		"clusterName": containerTable.State.ClusterName,
		"podId":       containerTable.State.PodID,
	}
}

func parseContainerStatus(ctx context.Context, state string) Status {
	color := ""
	switch state {
	case "running":
		color = "green"
	case "waiting":
		color = "steelblue"
	case "terminated":
		color = "red"
	}
	return Status{
		RenderType: "tagsRow",
		Size:       "default",
		Value: StatusValue{
			Label: cputil.I18n(ctx, state),
			Color: color,
		},
	}
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "containerTable", func() servicehub.Provider {
		return &ContainerTable{}
	})
}
