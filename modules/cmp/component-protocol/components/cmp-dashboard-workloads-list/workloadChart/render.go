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

package workloadChart

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("cmp-dashboard-workloads-list", "workloadChart", func() servicehub.Provider {
		return &ComponentWorkloadChart{}
	})
}

func (c *ComponentWorkloadChart) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	_ cptype.ComponentEvent, _ *cptype.GlobalStateData) error {
	if err := c.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen workloadChart component state, %v", err)
	}
	if err := c.SetComponentValue(ctx); err != nil {
		return fmt.Errorf("faield to set workloadChart component value, %v", err)
	}
	return nil
}

func (c *ComponentWorkloadChart) GenComponentState(component *cptype.Component) error {
	if component == nil || component.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(component.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", component.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	c.State = state
	return nil
}

func (c *ComponentWorkloadChart) SetComponentValue(ctx context.Context) error {
	c.Props.Option.Tooltip.Trigger = "axis"
	c.Props.Option.Tooltip.AxisPointer.Type = "shadow"
	c.Props.Option.Grid = Grid{
		Left:         "3%",
		Right:        "4%",
		Bottom:       "3%",
		Top:          "15%",
		ContainLabel: true,
	}
	c.Props.Option.Color = []string{
		"green", "red", "steelBlue", "maroon",
	}
	c.Props.Option.Legend.Data = []string{
		"Active", "Error", "Succeeded", "Failed",
	}
	c.Props.Option.XAxis.Type = "value"
	c.Props.Option.YAxis.Type = "category"
	c.Props.Option.YAxis.Data = []string{
		"CronJobs", "Jobs", "DaemonSets", "StatefulSets", "Deployments",
	}

	// deployment
	activeDeploy := c.State.Values.DeploymentsCount.Active
	errorDeploy := c.State.Values.DeploymentsCount.Error

	// daemonSet
	activeDs := c.State.Values.DaemonSetCount.Active
	errorDs := c.State.Values.DaemonSetCount.Error

	// statefulSet
	activeSs := c.State.Values.StatefulSetCount.Active
	errorSs := c.State.Values.StatefulSetCount.Error

	// job
	activeJob := c.State.Values.JobCount.Active
	succeededJob := c.State.Values.JobCount.Succeeded
	failedJob := c.State.Values.JobCount.Failed

	// cronjob
	activeCronJob := c.State.Values.CronJobCount.Active

	activeSeries := Series{
		Name:     cputil.I18n(ctx, "Active"),
		Type:     "bar",
		Stack:    "count",
		BarWidth: "50%",
		Data: []*int{
			&activeCronJob, &activeJob, &activeDs, &activeSs, &activeDeploy,
		},
	}

	errorSeries := Series{
		Name:     cputil.I18n(ctx, "Error"),
		Type:     "bar",
		Stack:    "count",
		BarWidth: "50%",
		Data: []*int{
			nil, nil, &errorDs, &errorSs, &errorDeploy,
		},
	}

	succeededSeries := Series{
		Name:     cputil.I18n(ctx, "Succeeded"),
		Type:     "bar",
		Stack:    "count",
		BarWidth: "50%",
		Data: []*int{
			nil, &succeededJob, nil, nil, nil,
		},
	}

	failedSeries := Series{
		Name:     cputil.I18n(ctx, "Failed"),
		Type:     "bar",
		Stack:    "count",
		BarWidth: "50%",
		Data: []*int{
			nil, &failedJob, nil, nil, nil,
		},
	}

	c.Props.Option.Series = []Series{
		activeSeries, errorSeries, succeededSeries, failedSeries,
	}
	return nil
}
