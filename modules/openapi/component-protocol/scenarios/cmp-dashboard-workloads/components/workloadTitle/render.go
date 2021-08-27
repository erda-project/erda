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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

func RenderCreator() protocol.CompRender {
	return &ComponentWorkloadTitle{}
}

func (w *ComponentWorkloadTitle) Render(ctx context.Context, component *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	if err := w.SetCtxBundle(ctx); err != nil {
		return fmt.Errorf("failed to set workloadTitle component ctx bundle, %v", err)
	}
	if err := w.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen workloadTitle component state, %v", err)
	}

	userID := w.ctxBdl.Identity.UserID
	orgID := w.ctxBdl.Identity.OrgID
	count := 0

	req := &apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        apistructs.K8SDeployment,
		ClusterName: w.State.ClusterName,
	}
	obj, err := w.ctxBdl.Bdl.ListSteveResource(req)
	if err != nil {
		return fmt.Errorf("failed to list deployments in render of workloadTitle component, %v", err)
	}
	count += len(obj.Slice("data"))

	req.Type = apistructs.K8SDaemonSet
	obj, err = w.ctxBdl.Bdl.ListSteveResource(req)
	if err != nil {
		return fmt.Errorf("failed to list daemonsets in render of workloadTitle component, %v", err)
	}
	count += len(obj.Slice("data"))

	req.Type = apistructs.K8SStatefulSet
	obj, err = w.ctxBdl.Bdl.ListSteveResource(req)
	if err != nil {
		return fmt.Errorf("failed to list statefulSets in render of workloadTitle component, %v", err)
	}
	count += len(obj.Slice("data"))

	req.Type = apistructs.K8SJob
	obj, err = w.ctxBdl.Bdl.ListSteveResource(req)
	if err != nil {
		return fmt.Errorf("failed to list jobs in render of workloadTitle component, %v", err)
	}
	count += len(obj.Slice("data"))

	req.Type = apistructs.K8SCronJob
	obj, err = w.ctxBdl.Bdl.ListSteveResource(req)
	if err != nil {
		return fmt.Errorf("failed to list cronJobs in render of workloadTitle component, %v", err)
	}
	count += len(obj.Slice("data"))

	w.Props.Title = fmt.Sprintf("Total workloads: %d", count)
	w.Props.Size = "small"
	return nil
}

func (w *ComponentWorkloadTitle) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil {
		return errors.New("bundle in context can not be empty")
	}
	w.ctxBdl = bdl
	return nil
}

func (w *ComponentWorkloadTitle) GenComponentState(component *apistructs.Component) error {
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
