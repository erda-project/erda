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

package workloadInfo

import (
	"context"
	"encoding/json"
	"fmt"

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
	base.InitProviderWithCreator("cmp-dashboard-workload-detail", "workloadInfo", func() servicehub.Provider {
		return &ComponentWorkloadInfo{}
	})
}

func (i *ComponentWorkloadInfo) Render(ctx context.Context, component *cptype.Component, _ cptype.Scenario,
	event cptype.ComponentEvent, _ *cptype.GlobalStateData) error {
	i.InitComponent(ctx)
	if err := i.GenComponentState(component); err != nil {
		return fmt.Errorf("failed to gen workloadInfo component state, %v", err)
	}
	if err := i.SetComponentValue(ctx); err != nil {
		return fmt.Errorf("failed to set workloadInfo component value, %v", err)
	}
	return nil
}

func (i *ComponentWorkloadInfo) InitComponent(ctx context.Context) {
	bdl := ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	i.bdl = bdl
	sdk := cputil.SDK(ctx)
	i.sdk = sdk
}

func (i *ComponentWorkloadInfo) GenComponentState(c *cptype.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var infoState State
	data, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(data, &infoState); err != nil {
		return err
	}
	i.State = infoState
	return nil
}

func (i *ComponentWorkloadInfo) SetComponentValue(ctx context.Context) error {
	kind, namespace, name, err := cmpcputil.ParseWorkloadID(i.State.WorkloadID)
	if err != nil {
		return err
	}
	userID := i.sdk.Identity.UserID
	orgID := i.sdk.Identity.OrgID

	req := apistructs.SteveRequest{
		UserID:      userID,
		OrgID:       orgID,
		Type:        kind,
		ClusterName: i.State.ClusterName,
		Name:        name,
		Namespace:   namespace,
	}
	obj, err := i.bdl.GetSteveResource(&req)
	if err != nil {
		return err
	}

	age, images, err := cmpcputil.GetWorkloadAgeAndImage(obj)
	if err != nil {
		return err
	}

	labels := obj.Map("metadata", "labels")
	labelTag := make([]Tag, 0)
	for key, value := range labels {
		labelTag = append(labelTag, Tag{Label: fmt.Sprintf("%s=%v", key, value)})
	}
	annotations := obj.Map("metadata", "annotations")
	annotationTag := make([]Tag, 0)
	for key, value := range annotations {
		annotationTag = append(annotationTag, Tag{Label: fmt.Sprintf("%s=%v", key, value)})
	}
	data := DataInData{
		Namespace:   namespace,
		Age:         age,
		Images:      images,
		Labels:      labelTag,
		Annotations: annotationTag,
	}
	i.Data.Data = data

	i.Props.ColumnNum = 4
	i.Props.Fields = []Field{
		{
			Label:    cputil.I18n(ctx, "namespace"),
			ValueKey: "namespace",
		},
		{
			Label:    cputil.I18n(ctx, "age"),
			ValueKey: "age",
		},
		{
			Label:      cputil.I18n(ctx, "images"),
			ValueKey:   "images",
			RenderType: "copyText",
		},
		{
			Label:      cputil.I18n(ctx, "labels"),
			ValueKey:   "labels",
			RenderType: "tagsRow",
			SpaceNum:   2,
		},
		{
			Label:      cputil.I18n(ctx, "annotations"),
			ValueKey:   "annotations",
			RenderType: "tagsRow",
			SpaceNum:   2,
		},
	}
	return nil
}
