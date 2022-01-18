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

package fileSearch

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
)

type ComponentAction struct {
	Props map[string]interface{} `json:"props"`
}

func (i *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	i.Props = map[string]interface{}{}
	i.Props["disabled"] = true
	i.Props["placeholder"] = "搜索(敬请期待)"
	c.Props = i.Props
	return nil
}

func init() {
	base.InitProviderWithCreator("auto-test-scenes", "fileSearch",
		func() servicehub.Provider { return &ComponentAction{} })
}
