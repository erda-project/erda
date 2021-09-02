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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	cmpcputil "github.com/erda-project/erda/modules/cmp/component-protocol/cputil"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("cmp-dashboard-workloads-list", "workloadChart", func() servicehub.Provider {
		return &ComponentWorkloadChart{}
	})
}

func (c *ComponentWorkloadChart) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	_ cptype.ComponentEvent, _ *cptype.GlobalStateData) error {
	c.InitComponent(ctx)
	if err := c.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen workloadChart component state, %v", err)
	}
	if err := c.SetComponentValue(); err != nil {
		return fmt.Errorf("faield to set workloadChart component value, %v", err)
	}
	return nil
}

func (c *ComponentWorkloadChart) InitComponent(ctx context.Context) {
	bdl := ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	c.bdl = bdl
	sdk := cputil.SDK(ctx)
	c.sdk = sdk
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

func (c *ComponentWorkloadChart) SetComponentValue() error {
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
		"green", "red", "steelBlue", "red",
	}
	c.Props.Option.Legend.Data = []string{
		"Active", "Error", "Succeeded", "Failed",
	}
	c.Props.Option.XAxis.Type = "value"
	c.Props.Option.YAxis.Type = "category"
	c.Props.Option.YAxis.Data = []string{
		"CronJobs", "Jobs", "DaemonSets", "StatefulSets", "Deployments",
	}

	userID := c.sdk.Identity.UserID
	orgID := c.sdk.Identity.OrgID
	req := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		ClusterName: c.State.ClusterName,
	}

	// deployment
	var activeDeploy, errorDeploy int
	req.Type = apistructs.K8SDeployment
	obj, err := c.bdl.ListSteveResource(&req)
	if err != nil {
		return err
	}
	list := obj.Slice("data")
	for _, obj := range list {
		status, _, err := cmpcputil.ParseWorkloadStatus(obj)
		if err != nil {
			logrus.Error(err)
			continue
		}
		if status == "Active" {
			activeDeploy++
		} else {
			errorDeploy++
		}
	}

	// daemonSet
	var activeDs, errorDs int
	req.Type = apistructs.K8SDaemonSet
	obj, err = c.bdl.ListSteveResource(&req)
	if err != nil {
		return err
	}
	list = obj.Slice("data")
	for _, obj := range list {
		status, _, err := cmpcputil.ParseWorkloadStatus(obj)
		if err != nil {
			logrus.Error(err)
			continue
		}
		if status == "Active" {
			activeDs++
		} else {
			errorDs++
		}
	}

	// statefulSet
	var activeSs, errorSs int
	req.Type = apistructs.K8SStatefulSet
	obj, err = c.bdl.ListSteveResource(&req)
	if err != nil {
		return err
	}
	list = obj.Slice("data")
	for _, obj := range list {
		status, _, err := cmpcputil.ParseWorkloadStatus(obj)
		if err != nil {
			logrus.Error(err)
			continue
		}
		if status == "Active" {
			activeSs++
		} else {
			errorSs++
		}
	}

	// job
	var activeJob, succeededJob, failedJob int
	req.Type = apistructs.K8SJob
	obj, err = c.bdl.ListSteveResource(&req)
	if err != nil {
		return err
	}
	list = obj.Slice("data")
	for _, obj := range list {
		status, _, err := cmpcputil.ParseWorkloadStatus(obj)
		if err != nil {
			logrus.Error(err)
			continue
		}
		if status == "Failed" {
			failedJob++
		} else if status != "Active" {
			activeJob++
		} else {
			succeededJob++
		}
	}

	// cronjob
	var activeCronJob int
	req.Type = apistructs.K8SCronJob
	obj, err = c.bdl.ListSteveResource(&req)
	if err != nil {
		return err
	}
	list = obj.Slice("data")
	activeCronJob = len(list)

	activeSeries := Series{
		Name:     "Active",
		Type:     "bar",
		Stack:    "count",
		BarWidth: "50%",
		Label: Label{
			Show:     true,
			Position: "insideRight",
		},
		Data: []*int{
			&activeCronJob, &activeJob, &activeDs, &activeSs, &activeDeploy,
		},
	}

	errorSeries := Series{
		Name:     "Error",
		Type:     "bar",
		Stack:    "count",
		BarWidth: "50%",
		Label: Label{
			Show:     true,
			Position: "insideRight",
		},
		Data: []*int{
			nil, nil, &errorDs, &errorSs, &errorDeploy,
		},
	}

	succeededSeries := Series{
		Name:     "Succeeded",
		Type:     "bar",
		Stack:    "count",
		BarWidth: "50%",
		Label: Label{
			Show:     true,
			Position: "insideRight",
		},
		Data: []*int{
			nil, &succeededJob, nil, nil, nil,
		},
	}

	failedSeries := Series{
		Name:     "Failed",
		Type:     "bar",
		Stack:    "count",
		BarWidth: "50%",
		Label: Label{
			Show:     true,
			Position: "insideRight",
		},
		Data: []*int{
			nil, &failedJob, nil, nil, nil,
		},
	}

	c.Props.Option.Series = []Series{
		activeSeries, errorSeries, succeededSeries, failedSeries,
	}
	return nil
}
