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

package recordDrawer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-space-list/i18n"
)

type Props struct {
	Title string `json:"title"`
	Size  string `json:"size"`
}

type State struct {
	Visible bool `json:"visible"`
}

type RecordDrawer struct {
	ctxBdl protocol.ContextBundle

	Type  string `json:"type"`
	Props Props  `json:"props"`
	State State  `json:"state"`
}

func (r *RecordDrawer) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	r.ctxBdl = bdl
	return nil
}

func (r *RecordDrawer) GenComponentState(c *apistructs.Component) error {
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
	r.State = state
	return nil
}

func (r *RecordDrawer) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := r.SetCtxBundle(ctx); err != nil {
		return err
	}
	if err := r.GenComponentState(c); err != nil {
		return err
	}
	i18nLocale := r.ctxBdl.Bdl.GetLocale(r.ctxBdl.Locale)
	r.Type = "Drawer"
	r.Props.Title = i18nLocale.Get(i18n.I18nKeyImportExportTable)
	r.Props.Size = "xl"

	return nil
}

func RenderCreator() protocol.CompRender {
	return &RecordDrawer{}
}
