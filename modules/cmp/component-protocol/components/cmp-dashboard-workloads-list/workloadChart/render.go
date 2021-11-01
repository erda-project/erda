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

func (w *ComponentWorkloadChart) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	_ cptype.ComponentEvent, _ *cptype.GlobalStateData) error {
	if err := w.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen workloadChart component state, %v", err)
	}
	if err := w.SetComponentValue(ctx); err != nil {
		return fmt.Errorf("faield to set workloadChart component value, %v", err)
	}
	w.Transfer(component)
	return nil
}

func (w *ComponentWorkloadChart) GenComponentState(c *cptype.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	w.State = state
	return nil
}

func (w *ComponentWorkloadChart) SetComponentValue(ctx context.Context) error {
	w.Props.Option.Tooltip.Trigger = "axis"
	w.Props.Option.Tooltip.AxisPointer.Type = "shadow"
	w.Props.Option.Color = []string{
		"green", "red", "steelblue", "maroon",
	}
	w.Props.Option.Legend.Data = []string{
		cputil.I18n(ctx, "Active"), cputil.I18n(ctx, "Error"), cputil.I18n(ctx, "Succeeded"), cputil.I18n(ctx, "Failed"),
	}
	w.Props.Option.XAxis.Type = "value"
	w.Props.Option.YAxis.Type = "category"
	w.Props.Option.YAxis.Data = []string{
		"CronJobs", "Jobs", "DaemonSets", "StatefulSets", "Deployments",
	}

	// deployment
	activeDeploy := w.State.Values.DeploymentsCount.Active
	errorDeploy := w.State.Values.DeploymentsCount.Error

	// daemonSet
	activeDs := w.State.Values.DaemonSetCount.Active
	errorDs := w.State.Values.DaemonSetCount.Error

	// statefulSet
	activeSs := w.State.Values.StatefulSetCount.Active
	errorSs := w.State.Values.StatefulSetCount.Error

	// job
	activeJob := w.State.Values.JobCount.Active
	succeededJob := w.State.Values.JobCount.Succeeded
	failedJob := w.State.Values.JobCount.Failed

	// cronjob
	activeCronJob := w.State.Values.CronJobCount.Active

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

	w.Props.Option.Series = []Series{
		activeSeries, errorSeries, succeededSeries, failedSeries,
	}
	return nil
}

func (w *ComponentWorkloadChart) Transfer(component *cptype.Component) {
	component.Props = w.Props
	component.State = map[string]interface{}{
		"values": w.State.Values,
	}
}
