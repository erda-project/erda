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

package pipelineTabs

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/project-pipeline/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/providers/projectpipeline"
)

type (
	Tab struct {
		Type               string                 `json:"type"`
		Data               Data                   `json:"data"`
		State              State                  `json:"state"`
		Operations         map[string]interface{} `json:"operations"`
		InParams           InParams               `json:"-"`
		sdk                *cptype.SDK
		bdl                *bundle.Bundle
		gsHelper           *gshelper.GSHelper
		ProjectPipelineSvc *projectpipeline.ProjectPipelineService
	}
	Data struct {
		Options []Option `json:"options"`
	}
	Option struct {
		Label string `json:"label"`
		Value string `json:"value"`
	}
	State struct {
		Value string `json:"value"`
	}
	InParams struct {
		FrontendProjectID string `json:"projectId,omitempty"`
		FrontendAppID     string `json:"appId,omitempty"`
		FrontendUrlQuery  string `json:"issueFilter__urlQuery,omitempty"`

		ProjectID uint64 `json:"-"`
		AppID     uint64 `json:"-"`
	}
)

func (t *Tab) setInParams(ctx context.Context) error {
	b, err := json.Marshal(cputil.SDK(ctx).InParams)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &t.InParams); err != nil {
		return err
	}

	if t.InParams.FrontendProjectID != "" {
		t.InParams.ProjectID, err = strconv.ParseUint(t.InParams.FrontendProjectID, 10, 64)
		if err != nil {
			return err
		}
	}
	if t.InParams.FrontendAppID != "" {
		t.InParams.AppID, err = strconv.ParseUint(t.InParams.FrontendAppID, 10, 64)
		if err != nil {
			return err
		}
	}

	return err
}

type Num struct {
	MinePipelineNum    uint64 `json:"minePipelineNum"`
	PrimaryPipelineNum uint64 `json:"primaryPipelineNum"`
	AllPipelineNum     uint64 `json:"allPipelineNum"`
}

func (t *Tab) SetType() {
	t.Type = "RadioTabs"
}

func (t *Tab) SetOperations(activeKey string) {
	t.Operations = map[string]interface{}{
		"onChange": map[string]interface{}{
			"fillMeta": "",
			"key":      "ChangeViewType",
			"meta": map[string]interface{}{
				"activeKey": activeKey,
			},
			"reload": true,
		},
	}
}

func (t *Tab) SetData(ctx context.Context, num Num) {
	t.Data = Data{Options: []Option{
		{
			Label: cputil.I18n(ctx, "allPipeline") + fmt.Sprintf("(%d)", num.AllPipelineNum),
			Value: common.AllState.String(),
		},
		{
			Label: cputil.I18n(ctx, "minePipeline") + fmt.Sprintf("(%d)", num.MinePipelineNum),
			Value: common.MineState.String(),
		},
		{
			Label: cputil.I18n(ctx, "primaryPipeline") + fmt.Sprintf("(%d)", num.PrimaryPipelineNum),
			Value: common.PrimaryState.String(),
		},
	}}
}

func (t *Tab) SetState(s string) {
	t.State = State{Value: s}
}

func (t *Tab) SetToProtocolComponent(c *cptype.Component) error {
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, &c)
}

func (t *Tab) InitFromProtocol(ctx context.Context, c *cptype.Component, gs *cptype.GlobalStateData) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(b, t); err != nil {
		return err
	}
	if err = t.setInParams(ctx); err != nil {
		return err
	}

	t.gsHelper = gshelper.NewGSHelper(gs)
	t.sdk = cputil.SDK(ctx)
	t.bdl = t.sdk.Ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	t.ProjectPipelineSvc = ctx.Value(types.ProjectPipelineService).(*projectpipeline.ProjectPipelineService)
	return nil
}
