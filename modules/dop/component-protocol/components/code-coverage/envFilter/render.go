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

package envFilter

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentAction struct {
	base.DefaultProvider
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {

	if c.State == nil {
		c.State = map[string]interface{}{}
	}

	var defaultEnv string
	if c.State["value"] == nil {
		defaultEnv = apistructs.DevEnv
	} else {
		defaultEnv = c.State["value"].(string)
	}

	c.Type = "Radio"
	c.Operations = map[string]interface{}{
		"onChange": map[string]interface{}{
			"key":    "changeEnv",
			"reload": true,
		},
	}
	c.State = map[string]interface{}{
		"value": defaultEnv,
	}
	c.Props = map[string]interface{}{
		"buttonStyle": "solid",
		"radioType":   "button",
		"options": []map[string]interface{}{
			{
				"key":  apistructs.DevEnv,
				"text": "开发环境",
			},
			{
				"key":  apistructs.TestEnv,
				"text": "测试环境",
			},
			{
				"key":  apistructs.StagingEnv,
				"text": "预发环境",
			},
			{
				"key":  apistructs.ProdEnv,
				"text": "生产环境",
			},
		},
	}

	return nil
}

func init() {
	base.InitProviderWithCreator("code-coverage", "envFilter", func() servicehub.Provider {
		return &ComponentAction{}
	})
}
