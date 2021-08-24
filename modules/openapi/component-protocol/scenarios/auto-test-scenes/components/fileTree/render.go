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

package fileTree

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// SetCtxBundle 设置bundle
func (i *ComponentFileTree) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		err := fmt.Errorf("invalie context bundle")
		return err
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	i.CtxBdl = b
	return nil
}

// GenComponentState 获取state
func (i *ComponentFileTree) GenComponentState(c *apistructs.Component) error {
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

func (i *ComponentFileTree) marshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(i.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	var data apistructs.ComponentData = map[string]interface{}{}
	data["treeData"] = i.Data
	c.Data = data
	c.State = state
	// c.Type = i.Type
	return nil
}

func (a *ComponentFileTree) unmarshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(c.State)
	if err != nil {
		return err
	}
	var state State
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	var data []SceneSet
	dataJson, err := json.Marshal(c.Data["treeData"])
	if err != nil {
		return err
	}
	err = json.Unmarshal(dataJson, &data)
	if err != nil {
		return err
	}
	a.State = state
	// a.Type = c.Type
	a.Data = data
	return nil
}

func (i *ComponentFileTree) RenderProtocol(c *apistructs.Component, g *apistructs.GlobalStateData) {
	if c.Data == nil {
		d := make(apistructs.ComponentData)
		c.Data = d
	}
	if c.Operations == nil {
		d := make(apistructs.ComponentOps)
		c.Operations = d
	}
	c.Props = i.Props
	c.Operations = i.Operations
	(*c).Data["treeData"] = i.Data
}

func (i *ComponentFileTree) onClickFolderTable() error {
	i.State.IsClickFolderTable = false
	i.State.IsClickScene = true
	id := i.State.ClickFolderTableSceneID

	var req apistructs.AutotestSceneRequest
	req.SceneID = id
	req.UserID = i.CtxBdl.Identity.UserID
	scene, err := i.CtxBdl.Bdl.GetAutoTestScene(req)
	if err != nil {
		return err
	}
	if scene.RefSetID != 0 {
		i.SetSceneSetClick(int(scene.RefSetID))
	} else {
		i.State.SceneId = id
		i.State.SceneId__urlQuery = strconv.FormatUint(id, 10)
		i.State.SelectedKeys = []string{"scene-" + i.State.SceneId__urlQuery}
	}
	return nil
}

func (i *ComponentFileTree) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	if event.Operation != apistructs.InitializeOperation && event.Operation != apistructs.RenderingOperation {
		err = i.unmarshal(c)
		if err != nil {
			return err
		}
	}
	i.State.IsClickScene = false

	defer func() {
		fail := i.marshal(c)
		if err == nil && fail != nil {
			err = fail
		}
	}()

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	err = i.SetCtxBundle(bdl)
	if err != nil {
		return err
	}
	err = i.GenComponentState(c)
	if err != nil {
		return err
	}

	var drag = Operation{
		Key:    "DragSceneSet",
		Reload: true,
	}

	i.Operations = map[string]interface{}{}
	i.Operations["drag"] = drag

	i.Props = map[string]interface{}{}
	i.Props["draggable"] = true

	inParamsBytes, err := json.Marshal(i.CtxBdl.InParams)
	if err != nil {
		return fmt.Errorf("failed to marshal inParams, inParams:%+v, err:%v", i.CtxBdl.InParams, err)
	}

	var inParams InParams
	if err := json.Unmarshal(inParamsBytes, &inParams); err != nil {
		return err
	}

	i.State.FormVisible = false
	i.State.ActionType = ""
	// i.State.SceneSetKey = 0
	// sceneId := fmt.Sprintf("%v", bdl.InParams["sceneId__urlQuery"])

	switch event.Operation {
	case apistructs.InitializeOperation:
		if inParams.SetID != "" {
			if err := i.RenderFileTree(inParams); err != nil {
				return err
			}
		} else {
			if err := i.RenderSceneSets(inParams); err != nil {
				return err
			}
		}
	case apistructs.RenderingOperation:
		if i.State.IsClickFolderTable {
			if err := i.onClickFolderTable(); err != nil {
				return err
			}
		}
		if err := i.RenderSceneSets(inParams); err != nil {
			return err
		}
	case apistructs.ExpandSceneSetOperationKey:
		if err := i.RenderExpandSceneSet(event); err != nil {
			return err
		}
	case apistructs.ClickSceneSetOperationKey:
		if err := i.RenderClickSceneSet(event); err != nil {
			return err
		}
	case apistructs.ClickSceneOperationKey:
		if err := i.RenderClickScene(event); err != nil {
			return err
		}
		i.State.IsClickScene = true
	case apistructs.AddSceneOperationKey:
		if err := i.RenderAddScene(event); err != nil {
			return err
		}
	case apistructs.UpdateSceneSetOperationKey:
		if err := i.RenderUpdateSceneSet(event); err != nil {
			return err
		}
	case apistructs.UpdateSceneOperationKey:
		if err := i.RenderUpdateScene(event); err != nil {
			return err
		}
	case apistructs.DeleteSceneSetOperationKey:
		if err := i.RenderDeleteSceneSet(event, inParams); err != nil {
			return err
		}
		if err := i.RenderSceneSets(inParams); err != nil {
			return err
		}
	case apistructs.DeleteSceneOperationKey:
		if err := i.RenderDeleteScene(event); err != nil {
			return err
		}
		if err := i.RenderSceneSets(inParams); err != nil {
			return err
		}
	case apistructs.DragSceneSetOperationKey:
		if err := i.RenderDragHelper(inParams); err != nil {
			return err
		}
		if err := i.RenderSceneSets(inParams); err != nil {
			return err
		}
	case apistructs.CopySceneOperationKey:
		if err := i.RenderCopyScene(inParams, event); err != nil {
			return err
		}
		if err := i.RenderSceneSets(inParams); err != nil {
			return err
		}
	case apistructs.RefSceneSetOperationKey:
		if err := i.RenderRefSceneSet(event); err != nil {
			return err
		}
	}
	i.RenderProtocol(c, gs)
	return
}

func RenderCreator() protocol.CompRender {
	return &ComponentFileTree{}
}
