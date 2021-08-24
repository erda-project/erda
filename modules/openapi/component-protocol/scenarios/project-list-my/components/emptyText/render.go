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

package emptyText

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/project-list-my/i18n"
)

// GenComponentState 获取state
func (i *ComponentText) GenComponentState(c *apistructs.Component) error {
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

func (i *ComponentText) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	i.ctxBdl = bdl
	return nil
}

func (i *ComponentText) Render(ctx context.Context, c *apistructs.Component, _ apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) (err error) {
	if err := i.GenComponentState(c); err != nil {
		return err
	}
	if err := i.SetCtxBundle(ctx); err != nil {
		return err
	}

	// 如果list组件的数据不为空，则直接返回
	if !i.State.IsEmpty {
		c.Props = map[string]interface{}{"visible": false}
		return nil
	}

	i18nLocale := i.ctxBdl.Bdl.GetLocale(i.ctxBdl.Locale)
	i.Props = Props{
		Visible:    true,
		RenderType: "linkText",
		StyleConfig: StyleConfig{
			FontSize:   16,
			LineHeight: 24,
		},
		Value: map[string]interface{}{
			"text": []interface{}{
				i18nLocale.Get(i18n.I18nProjectNotJoined),
			},
		},
	}

	c.Props = i.Props
	return
}

func RenderCreator() protocol.CompRender {
	return &ComponentText{}
}
