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

package recordButton

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
	Text  string `json:"text"`
	Type  string `json:"type"`
	Ghost bool   `json:"ghost"`
}

type State struct {
	Visible bool `json:"visible"`
}

type RecordButton struct {
	ctxBdl protocol.ContextBundle

	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	Operations map[string]interface{} `json:"operations"`
	State      State                  `json:"state"`
}

func (r *RecordButton) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	r.ctxBdl = bdl
	return nil
}

func (a *RecordButton) marshal(c *apistructs.Component) error {
	stateValue, err := json.Marshal(a.State)
	if err != nil {
		return err
	}
	var state map[string]interface{}
	err = json.Unmarshal(stateValue, &state)
	if err != nil {
		return err
	}

	propValue, err := json.Marshal(a.Props)
	if err != nil {
		return err
	}
	var props map[string]interface{}
	err = json.Unmarshal(propValue, &props)
	if err != nil {
		return err
	}

	c.State = state
	c.Type = a.Type
	c.Props = props
	return nil
}

func (r *RecordButton) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := r.SetCtxBundle(ctx); err != nil {
		return err
	}
	var visible bool
	switch event.Operation {
	case "autoRefresh":
		visible = true
	case "openRecord":
		visible = true
	}

	i18nLocale := r.ctxBdl.Bdl.GetLocale(r.ctxBdl.Locale)
	r.Type = "Button"
	r.Props.Text = i18nLocale.Get(i18n.I18nKeyImportExportRecord)
	r.Props.Type = "primary"
	r.Props.Ghost = true
	r.State.Visible = visible
	r.Operations = map[string]interface{}{
		"click": map[string]interface{}{
			"key":    "openRecord",
			"reload": true,
		},
	}
	if err := r.marshal(c); err != nil {
		return err
	}

	return nil
}

func RenderCreator() protocol.CompRender {
	return &RecordButton{}
}
