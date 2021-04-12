// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package fileInfo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// SetCtxBundle 设置bundle
func (i *ComponentFileInfo) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		err := fmt.Errorf("invalie context bundle")
		return err
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	i.ctxBdl = b
	return nil
}

// GenComponentState 获取state
func (i *ComponentFileInfo) GenComponentState(c *apistructs.Component) error {
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
	return nil
}

// GetReq 从InParams中获取请求参数
func (i ComponentFileInfo) GetReq() error {
	var inParams InParams
	cont, err := json.Marshal(i.ctxBdl.InParams)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", i.ctxBdl.InParams, err)
		return err
	}
	err = json.Unmarshal(cont, &inParams)
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return err
	}
	i.InParams = inParams
	return nil
}

func (i *ComponentFileInfo) RenderProtocol(c *apistructs.Component, g *apistructs.GlobalStateData) {
	if c.Data == nil {
		d := make(apistructs.ComponentData)
		c.Data = d
	}
	(*c).Data["data"] = i.Data
	c.Props = i.Props
}

func (i *ComponentFileInfo) GetInParams() error {
	var inParams InParams
	cont, err := json.Marshal(i.ctxBdl.InParams)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", i.ctxBdl.InParams, err)
		return err
	}
	err = json.Unmarshal(cont, &inParams)
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return err
	}
	i.InParams = inParams
	return nil
}

func (i *ComponentFileInfo) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err = i.SetCtxBundle(bdl); err != nil {
		return
	}
	if err = i.GenComponentState(c); err != nil {
		return
	}
	// TODO debug

	i.State.AutotestSceneRequest.UserID = i.ctxBdl.Identity.UserID
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
	rsp, err := i.ctxBdl.Bdl.GetAutoTestScene(i.State.AutotestSceneRequest)
	if err != nil {
		return err
	}

	// TODO 由于这里涉及旧数据迁移，用户信息可能有问题，所以err不返回
	createrName, err := i.ctxBdl.Bdl.GetCurrentUser(rsp.CreatorID)
	if err == nil {
		rsp.CreatorID = createrName.Nick
	}

	if rsp.UpdaterID != "" {
		updaterName, err := i.ctxBdl.Bdl.GetCurrentUser(rsp.UpdaterID)
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

func RenderCreator() protocol.CompRender {
	return &ComponentFileInfo{}
}
