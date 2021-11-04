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
	base.InitProviderWithCreator("addon-mysql-consumer", "filter",
		func() servicehub.Provider { return &comp{} })
}

func (f *comp) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	pg := common.LoadPageDataAttachment(ctx)
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

	c.State["conditions"] = []filter.PropCondition{
		{
			Key:       "account",
			Label:     "账号",
			EmptyText: "全部",
			Fixed:     true,
			Type:      filter.PropConditionTypeSelect,
			Options: func() (opt []filter.PropConditionOption) {
				for _, acc := range ac.Accounts {
					opt = append(opt, filter.PropConditionOption{
						Value: acc.Id,
						Label: acc.Username,
					})
				}
				return
			}(),
		},
		{
			Key:       "state",
			Label:     "是否切换中",
			EmptyText: "全部",
			Fixed:     true,
			Type:      filter.PropConditionTypeSelect,
			Options: []filter.PropConditionOption{
				{Value: "PRE", Label: "是"},
				{Value: "CUR", Label: "否"},
			},
		},
		{
			Key:       "app",
			Label:     "应用",
			EmptyText: "全部",
			Fixed:     true,
			Type:      filter.PropConditionTypeSelect,
			Options: func() (opt []filter.PropConditionOption) {
				set := map[string]struct{}{}
				for _, att := range ac.Attachments {
					set[att.AppId] = struct{}{}
				}
				for _, app := range ac.Apps {
					if _, ok := set[strutil.String(app.ID)]; ok {
						opt = append(opt, filter.PropConditionOption{
							Label: app.Name,
							Value: strutil.String(app.ID),
						})
					}
				}
				return
			}(),
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
