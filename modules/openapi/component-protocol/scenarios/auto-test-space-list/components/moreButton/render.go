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

package moreButton

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
	"github.com/erda-project/erda/modules/openapi/component-protocol/component_spec/button"
	"github.com/erda-project/erda/modules/openapi/component-protocol/scenarios/auto-test-space-list/i18n"
)

type MoreButton struct {
	ctxBdl protocol.ContextBundle

	Props button.Props `json:"props"`
	State State        `json:"state"`
}

type State struct {
	Visible bool `json:"visible"`
}

func (i *MoreButton) SetCtxBundle(ctx context.Context) error {
	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	if bdl.Bdl == nil || bdl.I18nPrinter == nil {
		return fmt.Errorf("invalid context bundle")
	}
	logrus.Infof("inParams:%+v, identity:%+v", bdl.InParams, bdl.Identity)
	i.ctxBdl = bdl
	return nil
}

func (i *MoreButton) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	if err := i.SetCtxBundle(ctx); err != nil {
		return err
	}

	if event.Operation == "autoRefresh" || event.Operation == "openRecord" {
		i.State.Visible = true
	}
	i18nLocale := i.ctxBdl.Bdl.GetLocale(i.ctxBdl.Locale)
	i.Props = button.Props{
		Text:  i.sdk.I18n("moreOperations"),
		Type:  "primary",
		Ghost: true,
		Menu: []button.MenuItem{
			{
				Key:  "import",
				Text: i18nLocale.Get(i18n.I18nKeyImport),
				Operations: map[string]interface{}{
					"click": map[string]interface{}{
						"key":    "import",
						"reload": false,
					},
				},
			},
			{
				Key:  "record",
				Text: i18nLocale.Get(i18n.I18nKeyImportExportRecord),
				Operations: map[string]interface{}{
					"click": map[string]interface{}{
						"key":    "openRecord",
						"reload": true,
					},
				},
			},
		},
	}
	return nil
}

func RenderCreator() protocol.CompRender {
	return &MoreButton{}
}
