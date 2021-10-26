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
	"fmt"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"

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

	workspace, ok := c.State["workspace"].(string)
	if !ok {
		return fmt.Errorf("workspace was empty")
	}

	erdaVersion := os.Getenv("DICE_VERSION")
	if erdaVersion == "" {
		erdaVersion = "1.4"
	}

	var disableTip []string
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

		var workspace = workspace
		hasAddon, err := svc.JudgeSourcecovAddon(projectId, uint64(orgIDInt), workspace)
		if err != nil {
			return err
		}
		if !hasAddon {

			disableTip = append(disableTip, "当前环境未启用覆盖率插件，请参考 ")
			disableTip = append(disableTip, "__change_cover_docs__")

			jacocoDisable = true
			err := svc.Cancel(apistructs.CodeCoverageCancelRequest{
				ProjectID: projectId,
				Workspace: workspace,
				IdentityInfo: apistructs.IdentityInfo{
					UserID:         sdk.Identity.UserID,
					InternalClient: sdk.Identity.InternalClient,
				},
			})
			if err != nil {
				logrus.Errorf("not have %v addon, cancel coverage plan error %v", code_coverage.SourcecovAddonName, err)
			}
		}

		c.State["disableSourcecov"] = jacocoDisable

		if !jacocoDisable {
			list, err := svc.ListCodeCoverageRecord(apistructs.CodeCoverageListRequest{
				ProjectID: projectId,
				PageNo:    1,
				PageSize:  1,
				Workspace: workspace,
			})
			if err != nil {
				return err
			}

			if len(list.List) > 0 {
				record := list.List[0]
				if record.Status == apistructs.RunningStatus.String() {
					disableTip = append(disableTip, "代码覆盖率统计进行中，开始和结束按钮不可用, 请等待(耗时取决于应用多少和大小)，手动刷新后查看结果")
				}

				if record.Status == apistructs.EndingStatus.String() {
					disableTip = append(disableTip, "代码覆盖率统计明细生成中，开始和结束按钮不可用, 请等待(耗时取决于应用多少和大小)，手动刷新后查看结果")
				}
			}
		}
	}
	c.Type = "Text"
	c.Name = "name"
	c.Props = map[string]interface{}{
		"renderType": "linkText",
		"styleConfig": map[string]interface{}{
			"color": "rgba(0, 0, 0, 0.4)",
		},
		"value": map[string]interface{}{
			"text": getToolTip(disableTip),
		},
	}
	c.Operations = map[string]interface{}{
		"gotoDoc": map[string]interface{}{
			"key":    "gotoDoc",
			"reload": false,
			"command": map[string]interface{}{
				"key":     "goto",
				"jumpOut": true,
				"target":  fmt.Sprintf("https://docs.erda.cloud/%v/manual/dop/examples/deploy/e2e-code-coverage.html", erdaVersion),
			},
		},
	}
	return nil
}

func getToolTip(tips []string) []interface{} {
	var results = []interface{}{}
	for _, ti := range tips {
		if ti == "__change_cover_docs__" {
			results = append(results, map[string]interface{}{
				"operationKey": "gotoDoc",
				"text":         "如何启用集成测试代码覆盖率统计",
				"styleConfig": map[string]interface{}{
					"color": "#6a549e",
				},
			})
		} else {
			results = append(results, ti)
		}
	}
	return results
}

func init() {
	base.InitProviderWithCreator("code-coverage", "errText", func() servicehub.Provider {
		return &ComponentAction{}
	})
}
