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

package head

import (
	"context"
	"fmt"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"

	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/component-protocol/types"
	"github.com/erda-project/erda/modules/dop/services/code_coverage"
)

type ComponentAction struct{}

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {

	svc := ctx.Value(types.CodeCoverageService).(*code_coverage.CodeCoverage)
	sdk := cputil.SDK(ctx)
	projectId := sdk.InParams["projectId"].(uint64)

	var disable = false
	var disabledTip string

	switch event.Operation.String() {
	case apistructs.ClickOperation.String():
		data, err := svc.ListCodeCoverageRecord(apistructs.CodeCoverageListRequest{
			PageSize:  1,
			PageNo:    1,
			ProjectID: projectId,
			Statuses: []apistructs.CodeCoverageExecStatus{
				apistructs.ReadyStatus,
			},
			IdentityInfo: apistructs.IdentityInfo{
				UserID: sdk.Identity.UserID,
			},
		})
		if err != nil {
			return err
		}
		if data.Total <= 0 {
			return fmt.Errorf("not find ready recode")
		}

		err = svc.End(apistructs.CodeCoverageUpdateRequest{
			ID: data.List[0].ID,
			IdentityInfo: apistructs.IdentityInfo{
				UserID: sdk.Identity.UserID,
			},
		})
		if err != nil {
			return err
		}

		disable = true
		disabledTip = "收集中"
	case apistructs.InitializeOperation.String(), apistructs.RenderingOperation.String():
		ok, err := svc.JudgeCanEnd(projectId)
		if err != nil {
			disabledTip = err.Error()
			disable = true
		}
		disable = ok
	}

	c.Type = "Button"
	c.Props = map[string]interface{}{
		"text": "结束",
		"type": "primary",
	}

	c.Operations = map[string]interface{}{
		"click": map[string]interface{}{
			"key":         "click",
			"reload":      true,
			"disabledTip": disabledTip,
			"disabled":    disable,
		},
	}
	return nil
}

func init() {
	base.InitProvider("code-coverage", "endButton")
}
