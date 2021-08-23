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
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
)

// SetCtxBundle 设置bundle
func (i *ComponentFilter) SetCtxBundle(b protocol.ContextBundle) error {
	if b.Bdl == nil || b.I18nPrinter == nil {
		err := fmt.Errorf("invalie context bundle")
		return err
	}
	logrus.Infof("inParams:%+v, identity:%+v", b.InParams, b.Identity)
	i.ctxBdl = b
	return nil
}

// GenComponentState 获取state
func (i *ComponentFilter) GenComponentState(c *apistructs.Component) error {
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

func (i *ComponentFilter) SetComponentValue() {
	i.Props = Props{
		Delay: 1000,
	}
	i.Operations = map[string]interface{}{
		apistructs.ListProjectFilterOperation.String(): Operations{
			Reload: true,
			Key:    apistructs.ListProjectFilterOperation.String(),
		},
	}
	i.State.Conditions = []StateConditions{
		{
			Key:         "title",
			Label:       "标题",
			EmptyText:   "全部",
			Fixed:       true,
			ShowIndex:   2,
			Placeholder: "搜索",
			Type:        "input",
		},
	}
}

// RenderProtocol 渲染组件
func (i *ComponentFilter) RenderProtocol(c *apistructs.Component, g *apistructs.GlobalStateData) error {
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

func (i *ComponentFilter) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if err = i.SetCtxBundle(bdl); err != nil {
		return
	}
	if err = i.GenComponentState(c); err != nil {
		return
	}

	i.State.IsFirstFilter = false
	if event.Operation == apistructs.ListProjectFilterOperation {
		i.State.IsFirstFilter = true
	}

	i.SetComponentValue()

	if err := i.RenderProtocol(c, gs); err != nil {
		return err
	}
	return
}

func RenderCreator() protocol.CompRender {
	return &ComponentFilter{}
}
