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
	"encoding/json"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
	"github.com/erda-project/erda/modules/openapi/component-protocol/pkg/type_conversion"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-plan-list/i18n"
)

type AutoTestPlanFilter struct{}

func RenderCreator() protocol.CompRender {
	return &AutoTestPlanFilter{}
}

type Value struct {
	Archive   []string `json:"archive"`
	Name      string   `json:"name"`
	Iteration []uint64 `json:"iteration"`
}

func (tpm *AutoTestPlanFilter) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	projectID, err := type_conversion.InterfaceToUint64(bdl.InParams["projectId"])
	if err != nil {
		return err
	}
	iterations, err := bdl.Bdl.ListProjectIterations(apistructs.IterationPagingRequest{
		PageNo:              1,
		PageSize:            999,
		ProjectID:           projectID,
		WithoutIssueSummary: true,
	}, "0")
	if err != nil {
		return err
	}
	if c.State == nil {
		c.State = make(map[string]interface{})
	}
	c.State["conditions"] = tpm.setConditions(bdl, iterations)
	if event.Operation.String() == "filter" {
		if _, ok := c.State["values"]; ok {
			fiterDataBytes, err := json.Marshal(c.State["values"])
			if err != nil {
				return err
			}
			var values Value
			if err := json.Unmarshal(fiterDataBytes, &values); err != nil {
				return err
			}

			c.State["name"] = values.Name
			if _, ok := c.State["archive"]; ok {
				c.State["archive"] = nil
			}
			if len(values.Archive) == 1 {
				c.State["archive"] = values.Archive[0] == "archived"
			}
			c.State["iteration"] = values.Iteration
		}
	} else {
		c.State["name"] = ""
		c.State["archive"] = false
		c.State["values"] = Value{
			Archive: []string{"inprogress"},
		}
		c.State["iteration"] = []uint64{}
	}

	return nil
}

func (tpm *AutoTestPlanFilter) setConditions(ctxBundle protocol.ContextBundle, iterations []apistructs.Iteration) []filter.PropCondition {
	i18nLocale := ctxBundle.Bdl.GetLocale(ctxBundle.Locale)
	return []filter.PropCondition{
		{
			Key:         "name",
			Label:       i18nLocale.Get(i18n.I18nKeyPlanName),
			Fixed:       true,
			Placeholder: i18nLocale.Get(i18n.I18nKeyPlanNameRegex),
			Type:        filter.PropConditionTypeInput,
		},
		{
			Key:         "archive",
			Label:       i18nLocale.Get(i18n.I18nKeyArchive),
			EmptyText:   i18nLocale.Get(i18n.I18nKeyAll),
			Fixed:       true,
			Placeholder: i18nLocale.Get(i18n.I18nKeyPlanNameRegex),
			Type:        filter.PropConditionTypeSelect,
			Options: []filter.PropConditionOption{
				{
					Label: i18nLocale.Get(i18n.I18nKeyInProgress),
					Value: "inprogress",
				},
				{
					Label: i18nLocale.Get(i18n.I18nKeyArchived),
					Value: "archived",
				},
			},
		},
		{
			EmptyText: i18nLocale.Get(i18n.I18nKeyAll),
			Fixed:     true,
			Key:       "iteration",
			Label:     i18nLocale.Get(i18n.I18nKeyIteration),
			Options: func() (opts []filter.PropConditionOption) {
				for _, itr := range iterations {
					opts = append(opts, filter.PropConditionOption{
						Label: itr.Title,
						Value: itr.ID,
					})
				}
				return
			}(),
			Type: filter.PropConditionTypeSelect,
		},
	}
}
