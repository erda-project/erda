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

package tomorrow

import (
	"context"
	"encoding/json"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/issue-dashboard/common"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentAction struct {
	base.DefaultProvider
	common.OverviewProps `json:"props,omitempty"`
	State                common.StatsState `json:"state,omitempty"`
}

func init() {
	base.InitProviderWithCreator("issue-dashboard", "tomorrow",
		func() servicehub.Provider { return &ComponentAction{} })
}

func (f *ComponentAction) InitFromProtocol(ctx context.Context, c *cptype.Component) error {
	// component 序列化
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, f); err != nil {
		return err
	}
	return nil
}

func (f *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	if err := f.InitFromProtocol(ctx, c); err != nil {
		return err
	}

	f.OverviewProps = common.OverviewProps{
		RenderType: "linkText",
		Value: common.OverviewValue{
			Direction: "col",
			Text: []common.OverviewText{
				{
					Text: f.State.Stats.Tomorrow,
					StyleConfig: common.StyleConfig{
						FontSize: 20,
						Bold:     true,
						Color:    "orange",
					},
				},
				{
					Text: "明日截止",
					StyleConfig: common.StyleConfig{
						Color: "text-desc",
					},
					Icon:    "help",
					IconTip: "不包含本日截止数据",
				},
			},
		},
	}
	c.Props = f.OverviewProps
	return nil
}
