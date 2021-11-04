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
	"github.com/erda-project/erda-infra/pkg/strutil"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/addon-mysql-account/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/filter"
)

type comp struct {
	base.DefaultProvider
}

func init() {
	base.InitProviderWithCreator("addon-mysql-account", "filter",
		func() servicehub.Provider { return &comp{} })
}

func (f *comp) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	pg := common.LoadPageDataAccount(ctx)
	ac, err := common.LoadAccountData(ctx)
	if err != nil {
		return err
	}

	if c.State == nil {
		c.State = map[string]interface{}{}
	}

	switch event.Operation {
	case cptype.InitializeOperation, cptype.RenderingOperation:
		// inParams -> values
		c.State["values"] = pg.FilterValues
	case "filter":
		// values -> inParams
		v := c.State["values"].(map[string]interface{})
		pg.FilterValues = v
		ft, err := common.ToBase64(v)
		if err != nil {
			return err
		}
		c.State["filter__urlQuery"] = ft
	}

	var userIDs []string
	for _, u := range ac.Accounts {
		userIDs = append(userIDs, u.Creator)
	}
	userIDs = strutil.DedupSlice(userIDs, true)
	var userOpt []filter.PropConditionOption
	for _, u := range userIDs {
		userOpt = append(userOpt, filter.PropConditionOption{
			Value: u,
			Label: u,
		})
	}

	c.State["conditions"] = []filter.PropCondition{
		{
			Key:       "status",
			Label:     "使用状态",
			EmptyText: "全部",
			Fixed:     true,
			Type:      filter.PropConditionTypeSelect,
			Options: []filter.PropConditionOption{
				{Label: "未被使用", Value: "NO"},
				{Label: "使用中", Value: "YES"},
			},
		},
		{
			Key:       "creator",
			Label:     "创建者",
			EmptyText: "全部",
			Fixed:     true,
			Type:      filter.PropConditionTypeSelect,
			Options:   userOpt,
		},
	}

	c.Props = map[string]interface{}{
		"delay": 1000,
	}
	c.Operations = map[string]interface{}{
		"filter": cptype.Operation{
			Key:    "filter",
			Reload: true,
		},
	}
	return nil
}
