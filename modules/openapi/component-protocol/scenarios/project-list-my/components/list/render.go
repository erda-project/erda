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

package list

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// SetCtxBundle 设置bundle
func (i *ComponentList) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		err := fmt.Errorf("invalie context bundle")
		return err
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	i.ctxBdl = b
	return nil
}

// GenComponentState 获取state
func (i *ComponentList) GenComponentState(c *apistructs.Component) error {
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

func (i *ComponentList) SetComponentValue() error {
	i.Props = Props{
		PageSizeOptions: []string{"10", "20", "50", "100"},
	}

	visible, err := i.CheckVisible() // true 展示list，隐藏emptyText / false 隐藏list，展示emptyText
	if err != nil {
		return err
	}
	if visible {
		i.Props.Visible = true
		i.State.IsEmpty = false
	} else {
		i.Props.Visible = false
		i.State.IsEmpty = true
	}

	opPageNo := Operation{
		Key:      apistructs.OnChangePageNoOperation.String(),
		Reload:   true,
		FillMeta: "pageNo",
	}
	opPageSize := Operation{
		Key:      apistructs.OnChangePageSizeOperation.String(),
		Reload:   true,
		FillMeta: "pageSize",
	}
	i.Operations = make(map[string]interface{})
	i.Operations[apistructs.OnChangePageNoOperation.String()] = opPageNo
	i.Operations[apistructs.OnChangePageSizeOperation.String()] = opPageSize
	return nil
}

// RenderProtocol 渲染组件
func (i *ComponentList) RenderProtocol(c *apistructs.Component, g *apistructs.GlobalStateData) error {
	if c.Data == nil {
		d := make(apistructs.ComponentData)
		c.Data = d
	}
	(*c).Data["list"] = i.Data.List

	stateValue, err := json.Marshal(i.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}
	c.State = state

	c.Props = i.Props
	c.Operations = i.Operations
	return nil
}

func (i *ComponentList) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err = i.SetCtxBundle(bdl); err != nil {
		return
	}
	if err = i.GenComponentState(c); err != nil {
		return
	}

	switch event.Operation {
	case apistructs.InitializeOperation:
		i.State.PageNo = 1
		i.State.PageSize = 20
		if err := i.RenderList(); err != nil {
			return err
		}
		if err := i.SetComponentValue(); err != nil {
			return err
		}
	case apistructs.RenderingOperation:
		// 如果触发搜索，跳转到第一页
		if i.State.IsFirstFilter {
			i.State.PageNo = 1
			i.State.IsFirstFilter = false
		}
		if err := i.RenderList(); err != nil {
			return err
		}
		if err := i.SetComponentValue(); err != nil {
			return err
		}
	case apistructs.OnChangePageSizeOperation:
		if err := i.RenderChangePageSize(event.OperationData); err != nil {
			return err
		}
		if err = i.RenderList(); err != nil {
			return err
		}
		if err := i.SetComponentValue(); err != nil {
			return err
		}
	case apistructs.OnChangePageNoOperation:
		if err := i.RenderChangePageNo(event.OperationData); err != nil {
			return err
		}
		if err = i.RenderList(); err != nil {
			return err
		}
		if err := i.SetComponentValue(); err != nil {
			return err
		}
	case apistructs.ListProjectExistOperationKey:
		err := i.RenderExist(event.OperationData)
		if err != nil {
			return err
		}
		err = i.RenderList()
		if err != nil {
			return err
		}
		if err := i.SetComponentValue(); err != nil {
			return err
		}
	}
	if err := i.RenderProtocol(c, gs); err != nil {
		return err
	}
	return
}

func RenderCreator() protocol.CompRender {
	return &ComponentList{}
}
