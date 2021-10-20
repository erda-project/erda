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

package chartFilter

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/gshelper"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common/stackhandlers"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

func init() {
	base.InitProviderWithCreator("issue-dashboard", "chartFilter",
		func() servicehub.Provider { return &ComponentFilter{} })
}

func (f *ComponentFilter) InitFromProtocol(ctx context.Context, c *cptype.Component) error {
	// component 序列化
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, f); err != nil {
		return err
	}

	// sdk
	f.sdk = cputil.SDK(ctx)

	return nil
}

func (f *ComponentFilter) SetToProtocolComponent(c *cptype.Component) error {
	b, err := json.Marshal(f)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	return nil
}

func (f *ComponentFilter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := f.InitFromProtocol(ctx, c); err != nil {
		return err
	}

	if err := f.InitDefaultOperation(c, gs); err != nil {
		return err
	}

	return f.SetToProtocolComponent(c)
}

func (f *ComponentFilter) InitDefaultOperation(c *cptype.Component, gs *cptype.GlobalStateData) error {
	if f.State.Values.Type == "" {
		f.State.Values.Type = stackhandlers.Priority
	}
	if f.State.FrontendChangedKey == "type" {
		f.State.Values.Value = nil
	}
	helper := gshelper.NewGSHelper(gs)
	handler := stackhandlers.NewStackRetriever(
		stackhandlers.WithIssueStateList(helper.GetIssueStateList()),
		stackhandlers.WithIssueStageList(helper.GetIssueStageList()),
	).GetRetriever(f.State.Values.Type)

	options := []filter.PropConditionOption{
		{
			Label: "优先级",
			Value: stackhandlers.Priority,
		},
		{
			Label: "复杂度",
			Value: stackhandlers.Complexity,
		},
		{
			Label: "严重程度",
			Value: stackhandlers.Severity,
		},
	}
	if c.Name == "stateVerticalBar" {
		options = append(options, filter.PropConditionOption{
			Label: "引入源",
			Value: stackhandlers.Stage,
		})
	} else {
		options = append(options, []filter.PropConditionOption{
			{
				Label: "状态",
				Value: stackhandlers.State,
			},
			{
				Label: "引入源",
				Value: stackhandlers.Stage,
			},
		}...)
	}
	f.State.Conditions = []filter.PropCondition{
		{
			EmptyText: "全部",
			Fixed:     true,
			Key:       "type",
			Label:     "类型",
			Options:   options,
			Required:  true,
			Type:      filter.PropConditionTypeSelect,
			CustomProps: map[string]interface{}{
				"mode": "single",
			},
		},
		{
			EmptyText: "全部",
			Fixed:     true,
			Key:       "value",
			Label:     "具体值",
			Options:   handler.GetFilterOptions(),
			Type:      filter.PropConditionTypeSelect,
		},
	}

	f.Props = filter.Props{
		Delay: 1000,
	}
	f.Operations = map[filter.OperationKey]filter.Operation{
		OperationKeyFilter: {
			Key:    OperationKeyFilter,
			Reload: true,
		},
	}

	return nil
}
