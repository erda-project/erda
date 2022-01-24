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
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"

	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-space-list/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-space-list/i18n"
)

type ComponentFilter struct {
	ctxBdl protocol.ContextBundle

	filter.CommonFilter
	State State `json:"state,omitempty"`
}

type State struct {
	Conditions []filter.PropCondition  `json:"conditions,omitempty"`
	Values     common.FilterConditions `json:"values,omitempty"`
}

func (i *ComponentFilter) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	i.ctxBdl = bdl
	return nil
}

func (i *ComponentFilter) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := i.SetCtxBundle(ctx); err != nil {
		return err
	}

	i18nLocale := i.ctxBdl.Bdl.GetLocale(i.ctxBdl.Locale)
	i.Props = filter.Props{Delay: 1000}
	i.Operations = map[filter.OperationKey]filter.Operation{
		"filter": {
			Key:    "filter",
			Reload: true,
		},
	}
	i.State.Conditions = []filter.PropCondition{
		{
			EmptyText: i18nLocale.Get("empty-filter-bookmark"),
			Fixed:     true,
			Key:       "order",
			Label:     i18nLocale.Get("sort"),
			Options: []filter.PropConditionOption{
				{
					Label: i18nLocale.Get("timeOrder"),
					Value: "updated_at",
				},
				{
					Label: i18nLocale.Get("timeReverse"),
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
			EmptyText: i18nLocale.Get("autotest.plan.all"),
			Fixed:     true,
			Key:       "archiveStatus",
			Label:     i18nLocale.Get("state"),
			Options: []filter.PropConditionOption{
				{
					Label: i18nLocale.Get(i18n.I18nKeyAutoTestSpaceInit),
					Value: apistructs.TestSpaceInit,
				},
				{
					Label: i18nLocale.Get(i18n.I18nKeyAutoTestSpaceInProgress),
					Value: apistructs.TestSpaceInProgress,
				},
				{
					Label: i18nLocale.Get(i18n.I18nKeyAutoTestSpaceCompleted),
					Value: apistructs.TestSpaceCompleted,
				},
			},
			Type: filter.PropConditionTypeSelect,
		},
		{
			EmptyText:   i18nLocale.Get("autotest.plan.all"),
			Fixed:       true,
			Key:         "spaceName",
			Label:       i18nLocale.Get("title"),
			Placeholder: i18nLocale.Get("searchByName"),
			Type:        filter.PropConditionTypeInput,
		},
	}

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentFilter{}
}
