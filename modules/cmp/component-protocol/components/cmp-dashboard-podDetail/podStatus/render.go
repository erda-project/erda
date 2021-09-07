package PodStatus

import (
	"context"
	"encoding/json"
	"fmt"
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

func (podStatus *PodStatus) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := podStatus.GenComponentState(c); err != nil {
		return err
	}
	sdk := cputil.SDK(ctx)
	bdl := ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)

	userID := sdk.Identity.UserID
	orgID := sdk.Identity.OrgID

	splits := strings.Split(podStatus.State.PodID, "_")
	if len(splits) != 2 {
		return fmt.Errorf("invalid pod id: %s", podStatus.State.PodID)
	}

	namespace, name := splits[0], splits[1]
	req := &apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        apistructs.K8SPod,
		ClusterName: podStatus.State.ClusterName,
		Name:        name,
		Namespace:   namespace,
	}

	obj, err := bdl.GetSteveResource(req)
	if err != nil {
		return err
	}

	fields := obj.StringSlice("metadata", "fields")
	if len(fields) != 9 {
		return fmt.Errorf("pod %s/%s has invalid fields length", namespace, name)
	}
	status := fields[2]
	color := ""
	switch status {
	case "Completed":
		color = "steelBlue"
	case "ContainerCreating":
		color = "orange"
	case "CrashLoopBackOff":
		color = "red"
	case "Error":
		color = "maroon"
	case "Evicted":
		color = "darkgoldenrod"
	case "ImagePullBackOff":
		color = "darksalmon"
	case "Pending":
		color = "teal"
	case "Running":
		color = "lightgreen"
	case "Terminating":
		color = "brown"
	}

	podStatus.Props = Props{
		StyleConfig: StyleConfig{Color: color},
		Value:       status,
	}
	return nil
}

func (podStatus *PodStatus) GenComponentState(component *cptype.Component) error {
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
	podStatus.State = state
	return nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "podStatus", func() servicehub.Provider {
		return &PodStatus{
			Type: "Text",
		}
	})
}
