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

package outPutForm

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// SetCtxBundle 设置bundle
func (i *ComponentOutPutForm) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		err := fmt.Errorf("invalie context bundle")
		return err
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	i.ctxBdl = b
	return nil
}

// GenComponentState 获取state
func (i *ComponentOutPutForm) GenComponentState(c *apistructs.Component) error {
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

func (i *ComponentOutPutForm) RenderProtocol(c *apistructs.Component, g *apistructs.GlobalStateData) {
	if c.Data == nil {
		d := make(apistructs.ComponentData)
		c.Data = d
	}
	if c.State == nil {
		d := make(apistructs.ComponentData)
		c.State = d
	}
	if c.Operations == nil {
		c.Operations = make(apistructs.ComponentOps)
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

func (i *ComponentOutPutForm) GetInParams() error {
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

func (i *ComponentOutPutForm) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err = i.SetCtxBundle(bdl)
	if err != nil {
		return
	}
	err = i.GenComponentState(c)
	if err != nil {
		return
	}
	if err = i.GetInParams(); err != nil {
		return err
	}

	i.Props = make(map[string]interface{})
	i.State.AutotestSceneRequest.UserID = i.ctxBdl.Identity.UserID
	i.State.AutotestSceneRequest.SetID = i.InParams.SceneSetID
	i.State.AutotestSceneRequest.SceneID = i.State.SceneId

	visible := make(map[string]interface{})
	visible["visible"] = true
	if i.State.AutotestSceneRequest.SceneID == 0 {
		visible["visible"] = false
		c.Props = visible
		return
	}
	i.Props = visible

	switch event.Operation {
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		err = i.RenderListOutPutForm()
		if err != nil {
			return err
		}
	case apistructs.AutoTestSceneInputUpdateOperationKey:
		err = i.RenderUpdateOutPutForm()
		if err != nil {
			return err
		}
		err = i.RenderListOutPutForm()
		if err != nil {
			return err
		}
	}
	i.RenderProtocol(c, gs)
	return
}

func RenderCreator() protocol.CompRender {
	return &ComponentOutPutForm{}
}
