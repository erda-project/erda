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

package addScenesButton

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cpregister/base"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/modules/dop/component-protocol/components/auto-test-scenes/common/gshelper"
)

const AddSceneOperationKey cptype.OperationKey = "AddScene"

func init() {
	base.InitProviderWithCreator("auto-test-scenes", "addScenesButton", func() servicehub.Provider {
		return &ComponentAction{}
	})
}

type ComponentAction struct {
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	h := gshelper.NewGSHelper(gs)
	switch event.Operation {
	case AddSceneOperationKey:
		c.State = map[string]interface{}{
			"actionType":    AddSceneOperationKey.String(),
			"visible":       true,
			"sceneSetKey":   h.GetGlobalSelectedSetID(),
			"isAddParallel": false,
		}
	case cptype.InitializeOperation, cptype.RenderingOperation:
		c.Type = "Button"
		c.Props = map[string]interface{}{
			"text": "+ " + cputil.I18n(ctx, "scene"),
		}
		c.Operations = map[string]interface{}{
			"click": map[string]interface{}{
				"key":    AddSceneOperationKey.String(),
				"reload": true,
			},
		}
	}
	return nil
}
