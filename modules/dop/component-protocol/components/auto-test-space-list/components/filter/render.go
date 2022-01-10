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

package filter

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"

	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/auto-test-space-list/common"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/auto-test-space-list/i18n"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

type ComponentFilter struct {
	sdk *cptype.SDK

	filter.CommonFilter
	State State `json:"state,omitempty"`
}

type State struct {
	Conditions []filter.PropCondition  `json:"conditions,omitempty"`
	Values     common.FilterConditions `json:"values,omitempty"`
}

func (i *ComponentFilter) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	i.sdk = cputil.SDK(ctx)
	i.Props = filter.Props{Delay: 1000}
	i.Operations = map[filter.OperationKey]filter.Operation{
		"filter": {
			Key:    "filter",
			Reload: true,
		},
	}
	i.State.Conditions = []filter.PropCondition{
		{
			EmptyText: "未选择",
			Fixed:     true,
			Key:       "order",
			Label:     "排序",
			Options: []filter.PropConditionOption{
				{
					Label: "按时间顺序",
					Value: "updated_at",
				},
				{
					Label: "按时间倒序",
					Value: "updated_at desc",
				},
			},
			Required: true,
			Type:     filter.PropConditionTypeSelect,
			CustomProps: map[string]interface{}{
				"mode": "single",
			},
		},
		{
			EmptyText: "全部",
			Fixed:     true,
			Key:       "archiveStatus",
			Label:     "状态",
			Options: []filter.PropConditionOption{
				{
					Label: i.sdk.I18n(i18n.I18nKeyAutoTestSpaceInit),
					Value: apistructs.TestSpaceInit,
				},
				{
					Label: i.sdk.I18n(i18n.I18nKeyAutoTestSpaceInProgress),
					Value: apistructs.TestSpaceInProgress,
				},
				{
					Label: i.sdk.I18n(i18n.I18nKeyAutoTestSpaceCompleted),
					Value: apistructs.TestSpaceCompleted,
				},
			},
			Type: filter.PropConditionTypeSelect,
		},
		{
			EmptyText:   "全部",
			Fixed:       true,
			Key:         "spaceName",
			Label:       "标题",
			Placeholder: "根据名称过滤",
			Type:        filter.PropConditionTypeInput,
		},
	}

	return nil
}

func init() {
	base.InitProviderWithCreator("auto-test-space-list", "filter",
		func() servicehub.Provider { return &ComponentFilter{} })
}
