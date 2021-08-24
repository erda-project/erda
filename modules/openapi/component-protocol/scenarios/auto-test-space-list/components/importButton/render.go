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

package importButton

import (
	"context"
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

type ImportButton struct {
	ctxBdl protocol.ContextBundle

	Type       string                 `json:"type"`
	Props      Props                  `json:"props"`
	Operations map[string]interface{} `json:"operations"`
}

func (i *ImportButton) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	i.ctxBdl = bdl
	return nil
}

func (i *ImportButton) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := i.SetCtxBundle(ctx); err != nil {
		return err
	}

	i18nLocale := i.ctxBdl.Bdl.GetLocale(i.ctxBdl.Locale)
	i.Type = "Button"
	i.Props.Text = i18nLocale.Get(i18n.I18nKeyImport)
	i.Props.Type = "primary"
	i.Props.Ghost = true
	i.Operations = map[string]interface{}{
		"click": map[string]interface{}{
			"reload": false,
		},
	}

	return nil
}

func RenderCreator() protocol.CompRender {
	return &ImportButton{}
}
