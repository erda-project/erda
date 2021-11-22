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

func (w *ComponentWorkloadChart) GenComponentState(component *cptype.Component) error {
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
		cputil.I18n(ctx, "Active"), cputil.I18n(ctx, "Abnormal"), cputil.I18n(ctx, "Succeeded"), cputil.I18n(ctx, "Failed"), cputil.I18n(ctx, "Updating"),
	}
	w.Props.Option.XAxis.Type = "value"
	w.Props.Option.YAxis.Type = "category"
	w.Props.Option.YAxis.Data = []string{
		"CronJobs", "Jobs", "DaemonSets", "StatefulSets", "Deployments",
	}

	// deployment
	activeDeploy := w.State.Values.DeploymentsCount.Active
	abnormalDeploy := w.State.Values.DeploymentsCount.Abnormal
	updatingDeploy := w.State.Values.DeploymentsCount.Updating

	// daemonSet
	activeDs := w.State.Values.DaemonSetCount.Active
	abnormalDs := w.State.Values.DaemonSetCount.Abnormal

	// statefulSet
	activeSs := w.State.Values.StatefulSetCount.Active
	abnormalSs := w.State.Values.StatefulSetCount.Abnormal

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
		Name:     cputil.I18n(ctx, "Abnormal"),
		Type:     "bar",
		Stack:    "count",
		BarWidth: "50%",
		Data: []*int{
			nil, nil, &abnormalDs, &abnormalSs, &abnormalDeploy,
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

	updatingSeries := Series{
		Name:     cputil.I18n(ctx, "Updating"),
		Type:     "bar",
		Stack:    "count",
		BarWidth: "50%",
		Data: []*int{
			nil, nil, nil, nil, &updatingDeploy,
		},
	}
	w.Props.Option.Series = []Series{
		activeSeries, errorSeries, succeededSeries, failedSeries, updatingSeries,
	}
	return nil
}

func (w *ComponentWorkloadChart) Transfer(c *cptype.Component) {
	c.Props = w.Props
	c.State = map[string]interface{}{
		"values": w.State.Values,
	}
}
