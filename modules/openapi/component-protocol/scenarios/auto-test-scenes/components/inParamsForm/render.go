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

package inParamsForm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// SetCtxBundle 设置bundle
func (i *ComponentInParamsForm) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		err := fmt.Errorf("invalie context bundle")
		return err
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	i.ctxBdl = b
	return nil
}

// GenComponentState 获取state
func (i *ComponentInParamsForm) GenComponentState(c *apistructs.Component) error {
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

func (i *ComponentInParamsForm) GetInparams() error {
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

func (i *ComponentInParamsForm) RenderProtocol(c *apistructs.Component, g *apistructs.GlobalStateData) {
	if c.Data == nil {
		d := make(apistructs.ComponentData)
		c.Data = d
	}
	if c.State == nil {
		d := make(apistructs.ComponentData)
		c.State = d
	}
	if c.Operations == nil {
		d := make(apistructs.ComponentOps)
		c.Operations = d
	}
	c.State["list"] = i.Data.List
	c.Props = i.Props
	c.Operations["save"] = OperationBaseInfo{
		Key:    "save",
		Reload: true,
	}
}

func GetOpsInfo(opsData interface{}) (*OpMetaInfo, error) {
	if opsData == nil {
		err := fmt.Errorf("empty operation data")
		return nil, err
	}
	var op OperationInfo
	cont, err := json.Marshal(opsData)
	if err != nil {
		logrus.Errorf("marshal inParams failed, content:%v, err:%v", opsData, err)
		return nil, err
	}
	err = json.Unmarshal(cont, &op)
	if err != nil {
		logrus.Errorf("unmarshal move out request failed, content:%v, err:%v", cont, err)
		return nil, err
	}
	meta := op.Meta
	return &meta, nil
}

func (i *ComponentInParamsForm) GetInParams() error {
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

func (i *ComponentInParamsForm) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err = i.SetCtxBundle(bdl); err != nil {
		return
	}
	if err = i.GenComponentState(c); err != nil {
		return
	}
	if err = i.GetInParams(); err != nil {
		logrus.Errorf("get filter request failed, content:%+v, err:%v", *gs, err)
		return
	}

	i.Props = make(map[string]interface{})

	i.State.AutotestSceneRequest.UserID = i.ctxBdl.Identity.UserID
	i.State.AutotestSceneRequest.SceneID = i.State.SceneId
	i.State.AutotestSceneRequest.SpaceID = i.InParams.SpaceId
	i.State.AutotestSceneRequest.SetID = i.State.SetId

	list := i.Filter()

	visible := make(map[string]interface{})
	visible["visible"] = true
	if i.State.AutotestSceneRequest.SceneID == 0 {
		visible["visible"] = false
		c.Props = visible
		return
	}
	i.Props = visible

	i.SetProps()
	switch event.Operation {
	case apistructs.InitializeOperation:
		err = i.RenderListInParamsForm()
		if err != nil {
			return err
		}
	case apistructs.RenderingOperation:
		err = i.RenderListInParamsForm()
		if err != nil {
			return err
		}
	case apistructs.AutoTestSceneInputUpdateOperationKey:
		err = i.RenderUpdateInParamsForm()
		if err != nil {
			return err
		}
		i.Data.List = list
	case apistructs.AutoTestSceneInputOnSelectOperationKey:
		err = i.RenderOnSelect(event.OperationData)
		if err != nil {
			return err
		}
		i.Data.List = list
	}
	i.RenderProtocol(c, gs)
	return
}

func RenderCreator() protocol.CompRender {
	return &ComponentInParamsForm{}
}
