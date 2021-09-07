package podTitle

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func (podTitle *PodTitle) Render(ctx context.Context, c *cptype.Component, s cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := podTitle.GenComponentState(c); err != nil {
		return err
	}

	splits := strings.Split(podTitle.State.PodID, "_")
	if len(splits) != 2 {
		return fmt.Errorf("invalid pod name: %s", podTitle.State.PodID)
	}
	name := splits[1]
	podTitle.Props.Title = fmt.Sprintf("Pod: %s", name)
	return nil
}

func (podTitle *PodTitle) GenComponentState(component *cptype.Component) error {
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
	podTitle.State = state
	return nil
}

func init() {
	base.InitProviderWithCreator("cmp-dashboard-podDetail", "podTitle", func() servicehub.Provider {
		return &PodTitle{Type: "Title"}
	})
}
