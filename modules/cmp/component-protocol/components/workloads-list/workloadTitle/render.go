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

package workloadTitle

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("workloads-list", "workloadTitle", func() servicehub.Provider {
		return &ComponentWorkloadTitle{}
	})
}

func (w *ComponentWorkloadTitle) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	_ cptype.ComponentEvent, _ *cptype.GlobalStateData) error {
	w.InitComponent(ctx)
	if err := w.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen workloadTitle component state, %v", err)
	}

	userID := w.sdk.Identity.UserID
	orgID := w.sdk.Identity.OrgID
	count := 0

	req := &apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        apistructs.K8SDeployment,
		ClusterName: w.State.ClusterName,
	}
	obj, err := w.bdl.ListSteveResource(req)
	if err != nil {
		return fmt.Errorf("failed to list deployments in render of workloadTitle component, %v", err)
	}
	count += len(obj.Slice("data"))

	req.Type = apistructs.K8SDaemonSet
	obj, err = w.bdl.ListSteveResource(req)
	if err != nil {
		return fmt.Errorf("failed to list daemonsets in render of workloadTitle component, %v", err)
	}
	count += len(obj.Slice("data"))

	req.Type = apistructs.K8SStatefulSet
	obj, err = w.bdl.ListSteveResource(req)
	if err != nil {
		return fmt.Errorf("failed to list statefulSets in render of workloadTitle component, %v", err)
	}
	count += len(obj.Slice("data"))

	req.Type = apistructs.K8SJob
	obj, err = w.bdl.ListSteveResource(req)
	if err != nil {
		return fmt.Errorf("failed to list jobs in render of workloadTitle component, %v", err)
	}
	count += len(obj.Slice("data"))

	req.Type = apistructs.K8SCronJob
	obj, err = w.bdl.ListSteveResource(req)
	if err != nil {
		return fmt.Errorf("failed to list cronJobs in render of workloadTitle component, %v", err)
	}
	count += len(obj.Slice("data"))

	w.Props.Title = fmt.Sprintf("Total workloads: %d", count)
	w.Props.Size = "small"
	return nil
}

func (w *ComponentWorkloadTitle) InitComponent(ctx context.Context) {
	bdl := ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	w.bdl = bdl
	sdk := cputil.SDK(ctx)
	w.sdk = sdk
}

func (w *ComponentWorkloadTitle) GenComponentState(component *cptype.Component) error {
	if component == nil || component.State == nil {
		return nil
	}
	var state State
	data, err := json.Marshal(component.State)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &state); err != nil {
		return err
	}
	w.State = state
	return nil
}
