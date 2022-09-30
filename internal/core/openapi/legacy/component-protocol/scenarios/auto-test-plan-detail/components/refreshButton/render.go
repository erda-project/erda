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

package refreshButton

import (
	"context"

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/internal/core/openapi/legacy/component-protocol"
	"github.com/erda-project/erda/internal/core/openapi/legacy/component-protocol/pkg/gshelper"
	"github.com/erda-project/erda/internal/core/openapi/legacy/component-protocol/scenarios/auto-test-plan-detail/i18n"
)

type ComponentAction struct{}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {

	bdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	i18nLocale := bdl.Bdl.GetLocale(bdl.Locale)
	switch event.Operation {
	case apistructs.ClickOperation:
		c.State = map[string]interface{}{
			"reloadScenesInfo": true,
		}
	case apistructs.InitializeOperation, apistructs.RenderingOperation:
		c.Type = "Button"
		c.Props = map[string]interface{}{
			"text":       i18nLocale.Get(i18n.I18nKeyRefresh),
			"prefixIcon": "shuaxin",
		}
		c.Operations = map[string]interface{}{
			"click": map[string]interface{}{
				"key":    "refresh",
				"reload": true,
			},
		}
	}
	gh := gshelper.NewGSHelper(gs)
	gh.ClearPipelineInfo()
	return nil
}

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
