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

package workloadStatus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/erda-project/erda/apistructs"
	workloads "github.com/erda-project/erda/modules/cmp/component-protocol/scenarios"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	workloaddetail "github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/cmp-dashboard-workload-detail"
)

func RenderCreator() protocol.CompRender {
	return &ComponentWorkloadStatus{}
}

func (s *ComponentWorkloadStatus) Render(ctx context.Context, component *apistructs.Component, _ apistructs.ComponentProtocolScenario,
	event apistructs.ComponentEvent, _ *apistructs.GlobalStateData) error {
	if err := s.SetCtxBundle(ctx); err != nil {
		return fmt.Errorf("failed to set workloadStatus component ctx bundle, %v", err)
	}
	if err := s.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen workloadStatue component state, %v", err)
	}
	if err := s.SetComponentValue(); err != nil {
		return fmt.Errorf("failed to set workoadStatus component value, %v", err)
	}
	return nil
}

func (s *ComponentWorkloadStatus) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil {
		return errors.New("bundle in context can not be empty")
	}
	s.ctxBdl = bdl
	return nil
}

func (s *ComponentWorkloadStatus) GenComponentState(component *apistructs.Component) error {
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
	s.State = state
	return nil
}

func (s *ComponentWorkloadStatus) SetComponentValue() error {
	kind, namespace, name, err := workloaddetail.ParseWorkloadID(s.State.WorkloadID)
	if err != nil {
		return err
	}

	userID := s.ctxBdl.Identity.UserID
	orgID := s.ctxBdl.Identity.OrgID
	req := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        apistructs.K8SResType(kind),
		ClusterName: s.State.ClusterName,
		Name:        name,
		Namespace:   namespace,
	}

	obj, err := s.ctxBdl.Bdl.GetSteveResource(&req)
	if err != nil {
		return err
	}

	status, color, err := workloads.ParseWorkloadStatus(obj)
	if err != nil {
		return err
	}
	s.Props.Value = status
	s.Props.StyleConfig.Color = color
	return nil
}
