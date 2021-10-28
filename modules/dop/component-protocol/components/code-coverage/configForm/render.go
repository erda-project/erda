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

package configForm

import (
	"context"
	"fmt"
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

	if c.State == nil {
		c.State = map[string]interface{}{}
	}

	workspace, ok := c.State["workspace"].(string)
	if !ok {
		return fmt.Errorf("workspace was empty")
	}

	switch event.Operation.String() {
	case "submit":
		fromData := c.State["formData"].(map[string]interface{})
		var saveSettingRequest = apistructs.SaveCodeCoverageSettingRequest{
			ProjectID:    projectId,
			Workspace:    workspace,
			MavenSetting: fromData["maven"].(string),
			Includes:     fromData["include"].(string),
			Excludes:     fromData["exclude"].(string),
			IdentityInfo: apistructs.IdentityInfo{
				UserID:         sdk.Identity.UserID,
				InternalClient: sdk.Identity.InternalClient,
			},
		}
		_, err := svc.SaveCodeCoverageSetting(saveSettingRequest)
		if err != nil {
			return err
		}
		c.State["showSettingModal"] = false
	case "cancel":
		c.State["showSettingModal"] = false
	}

	setting, err := svc.GetCodeCoverageSetting(projectId, workspace)
	if err != nil {
		return err
	}

	c.State["formData"] = map[string]interface{}{
		"maven":   setting.MavenSetting,
		"exclude": setting.Excludes,
		"include": setting.Includes,
	}

	c.Type = "Form"
	c.Operations = map[string]interface{}{
		"submit": map[string]interface{}{
			"key":    "submit",
			"reload": "true",
		},
		"cancel": map[string]interface{}{
			"key":    "cancel",
			"reload": "true",
		},
	}
	c.Props = map[string]interface{}{
		"fields": []map[string]interface{}{
			{
				"label":     "包含",
				"component": "input",
				"key":       "include",
				"componentProps": map[string]interface{}{
					"placeholder": "请输入包含表达式",
				},
			},
			{
				"label":     "不包含",
				"component": "input",
				"key":       "exclude",
				"componentProps": map[string]interface{}{
					"placeholder": "请输入不包含表达式",
				},
			},
			{
				"label":     "maven 设置",
				"component": "textarea",
				"key":       "maven",
				"componentProps": map[string]interface{}{
					"placeholder": "请输入maven 设置表达式",
				},
			},
		},
	}
	return nil
}

func init() {
	base.InitProviderWithCreator("code-coverage", "configForm", func() servicehub.Provider {
		return &ComponentAction{}
	})
}
