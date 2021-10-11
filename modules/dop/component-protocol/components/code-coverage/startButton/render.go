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

package startButton

import (
	"context"
	"strconv"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/services/code_coverage"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

type ComponentAction struct {
	base.DefaultProvider
}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {

	svc := ctx.Value(types.CodeCoverageService).(*code_coverage.CodeCoverage)
	sdk := cputil.SDK(ctx)
	projectIDStr := sdk.InParams["projectId"].(string)
	projectId, err := strconv.ParseUint(projectIDStr, 10, 64)
	if err != nil {
		return err
	}

	var disable = false

	switch event.Operation.String() {
	case apistructs.ClickOperation.String():

		err := svc.Start(apistructs.CodeCoverageStartRequest{
			ProjectID: projectId,
			IdentityInfo: apistructs.IdentityInfo{
				UserID: sdk.Identity.UserID,
			},
		})
		if err != nil {
			return err
		}

		disable = true
	case apistructs.InitializeOperation.String(), apistructs.RenderingOperation.String():
		err := svc.JudgeRunningRecordExist(projectId)
		if err != nil {
			disable = true
		}
	}

	c.Type = "Button"
	c.Props = map[string]interface{}{
		"text": "开始",
		"type": "primary",
	}

	c.Operations = map[string]interface{}{
		"click": map[string]interface{}{
			"key":         "click",
			"reload":      true,
			"disabledTip": "进行或明细生成中请等待",
			"disabled":    disable,
		},
	}
	return nil
}

func init() {
	base.InitProviderWithCreator("code-coverage", "startButton", func() servicehub.Provider {
		return &ComponentAction{}
	})
}
