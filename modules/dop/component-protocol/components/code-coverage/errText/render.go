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

package errText

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

	var disableTip string

	var disableEnd bool
	var disableStart bool
	var jacocoDisable bool

	switch event.Operation.String() {
	case apistructs.InitializeOperation.String(), apistructs.RenderingOperation.String():
		if c.State == nil {
			c.State = map[string]interface{}{}
		}

		orgIDInt, err := strconv.ParseInt(sdk.Identity.OrgID, 10, 64)
		if err != nil {
			return err
		}

		judgeApplication, err, message := svc.JudgeApplication(projectId, uint64(orgIDInt), sdk.Identity.UserID)
		if err != nil {
			return err
		}
		jacocoDisable = !judgeApplication
		if jacocoDisable && message != "" {
			disableTip = message
		}

		if c.State == nil {
			c.State = map[string]interface{}{}
		}

		c.State["judgeApplication"] = judgeApplication
		c.State["judgeApplicationMessage"] = message

		if !jacocoDisable {
			ok, err := svc.JudgeCanEnd(projectId)
			if err != nil {
				return err
			}
			disableEnd = !ok

			err = svc.JudgeRunningRecordExist(projectId)
			if err != nil {
				disableStart = true
			} else {
				disableStart = false
			}
		}
	}
	if c.State == nil {
		c.State = map[string]interface{}{}
	}

	if disableStart {
		if disableEnd {
			disableTip = "代码覆盖率统计明细生成中，开始和结束按钮不可用, 请等待(耗时取决于应用多少和大小)，手动刷新后查看结果"
		} else {
			disableTip = "代码覆盖率统计进行中，开始和结束按钮不可用, 请等待(耗时取决于应用多少和大小)，手动刷新后查看结果"
		}
	}

	c.Props = map[string]interface{}{
		"value": disableTip,
		"styleConfig": map[string]interface{}{
			"color": "red",
		},
	}
	return nil
}

func init() {
	base.InitProviderWithCreator("code-coverage", "errText", func() servicehub.Provider {
		return &ComponentAction{}
	})
}
