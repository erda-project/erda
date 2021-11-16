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

package fileInfo

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda/bundle"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

func init() {
	base.InitProviderWithCreator("auto-test-scenes", "fileInfo",
		func() servicehub.Provider { return &ComponentFileInfo{} })
}

// GenComponentState 获取state
func (i *ComponentFileInfo) GenComponentState(ctx context.Context, c *cptype.Component) error {
	if c == nil || c.State == nil {
		return nil
	}
	var state State
	cont, err := json.Marshal(c.State)
	if err != nil {
		logrus.Errorf("marshal component state failed, content:%v, err:%v", c.State, err)
		return err
	}
	err = json.Unmarshal(cont, &state)
	if err != nil {
		logrus.Errorf("unmarshal component state failed, content:%v, err:%v", cont, err)
		return err
	}
	i.State = state
	i.sdk = cputil.SDK(ctx)
	return nil
}

func (i *ComponentFileInfo) RenderProtocol(c *cptype.Component, g *cptype.GlobalStateData) {
	if c.Data == nil {
		d := make(cptype.ComponentData)
		c.Data = d
	}
	(*c).Data["data"] = i.Data
	c.Props = i.Props
}

func (i *ComponentFileInfo) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) (err error) {
	i.bdl = ctx.Value(types.GlobalCtxKeyBundle).(*bundle.Bundle)
	if err = i.GenComponentState(ctx, c); err != nil {
		return
	}
	// TODO debug

	i.State.AutotestSceneRequest.UserID = i.sdk.Identity.UserID
	i.State.AutotestSceneRequest.SceneID = i.State.SceneId
	i.State.AutotestSceneRequest.SetID = i.InParams.SceneSetID

	visible := make(map[string]interface{})
	visible["visible"] = true
	if i.State.AutotestSceneRequest.SceneID == 0 {
		visible["visible"] = false
		c.Props = visible
		return
	}
	i.Props = visible
	rsp, err := i.bdl.GetAutoTestScene(i.State.AutotestSceneRequest)
	if err != nil {
		return err
	}

	// TODO 由于这里涉及旧数据迁移，用户信息可能有问题，所以err不返回
	createrName, err := i.bdl.GetCurrentUser(rsp.CreatorID)
	if err == nil {
		rsp.CreatorID = createrName.Nick
	}

	if rsp.UpdaterID != "" {
		updaterName, err := i.bdl.GetCurrentUser(rsp.UpdaterID)
		if err == nil {
			rsp.UpdaterID = updaterName.Nick
		}
	}

	i.Data.AutoTestScene = *rsp
	i.Data.CreateATString = rsp.CreateAt.Format("2006-01-02 15:04:05")
	i.Data.UpdateATString = rsp.UpdateAt.Format("2006-01-02 15:04:05")

	i.Props["fields"] = []PropColumn{
		{
			Label:      "名称",
			ValueKey:   "name",
			RenderType: "ellipsis",
		},
		{
			Label:      "描述",
			ValueKey:   "description",
			RenderType: "ellipsis",
		},
		{
			Label:    "创建人",
			ValueKey: "creatorID",
		},
		{
			Label:    "创建时间",
			ValueKey: "createAtString",
		},
		{
			Label:    "更新人",
			ValueKey: "updaterID",
		},
		{
			Label:    "更新时间",
			ValueKey: "updateAtString",
		},
	}

	i.RenderProtocol(c, gs)
	return nil
}
