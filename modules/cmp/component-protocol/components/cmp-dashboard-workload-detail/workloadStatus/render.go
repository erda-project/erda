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
	"fmt"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	cputil2 "github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/cmp/component-protocol/cputil"
	"github.com/erda-project/erda/modules/cmp/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("cmp-dashboard-workload-detail", "workloadStatus", func() servicehub.Provider {
		return &ComponentWorkloadStatus{}
	})
}

func (s *ComponentWorkloadStatus) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, _ *cptype.GlobalStateData) error {
	s.SetCtxBundle(ctx)
	if err := s.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen workloadStatue component state, %v", err)
	}
	if err := s.SetComponentValue(); err != nil {
		return fmt.Errorf("failed to set workoadStatus component value, %v", err)
	}
	return nil
}

func (s *ComponentWorkloadStatus) SetCtxBundle(ctx context.Context) {
	bdl := ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	s.bdl = bdl
	sdk := cputil2.SDK(ctx)
	s.sdk = sdk
}

func (s *ComponentWorkloadStatus) GenComponentState(c *cptype.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	data, err := json.Marshal(c.State)
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
	kind, namespace, name, err := cputil.ParseWorkloadID(s.State.WorkloadID)
	if err != nil {
		return err
	}

	userID := s.sdk.Identity.UserID
	orgID := s.sdk.Identity.OrgID
	req := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        kind,
		ClusterName: s.State.ClusterName,
		Name:        name,
		Namespace:   namespace,
	}

	obj, err := s.bdl.GetSteveResource(&req)
	if err != nil {
		return err
	}

	status, color, err := cputil.ParseWorkloadStatus(obj)
	if err != nil {
		return err
	}
	s.Data.Labels.Label = s.sdk.I18n(status)
	s.Data.Labels.Color = color
	s.Props.Size = "default"
	return nil
}
