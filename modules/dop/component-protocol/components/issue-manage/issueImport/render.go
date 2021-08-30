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

package issueImport

import (
	"context"

	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/component-protocol/cptype"
	"github.com/erda-project/erda-infra/providers/component-protocol/utils/cputil"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/bdl"
	"github.com/erda-project/erda/modules/openapi/component-protocol/components/base"
)

// {
//         "size": "small",
//         "tooltip": "导入",
//         "prefixIcon": "import"
//     }
type IssueImportProps struct {
	Size       string `json:"size"`
	Tooltip    string `json:"tooltip"`
	PrefixIcon string `json:"prefixIcon"`
	Visible    bool   `json:"visible"`
}

type ComponentAction struct{ base.DefaultProvider }

func (ca *ComponentAction) Render(ctx context.Context, c *cptype.Component, scenario cptype.Scenario, event cptype.ComponentEvent, gs *cptype.GlobalStateData) error {
	sdk := cputil.SDK(ctx)
	c.Props = IssueImportProps{
		Size:       "small",
		Tooltip:    "导入",
		PrefixIcon: "import",
		Visible:    sdk.InParams["fixedIssueType"].(string) != "ALL",
	}

	isGuest, err := ca.CheckUserPermission(ctx)
	if err != nil {
		return err
	}

	c.Operations = map[string]interface{}{
		"click": struct {
			Reload   bool `json:"reload"`
			Disabled bool `json:"disabled"`
		}{Reload: false, Disabled: isGuest},
	}
	return nil
}

// GetUserPermission  check Guest permission
func (ca *ComponentAction) CheckUserPermission(ctx context.Context) (bool, error) {
	sdk := cputil.SDK(ctx)
	isGuest := false
	projectID := sdk.InParams["projectId"].(string)
	scopeRole, err := bdl.Bdl.ScopeRoleAccess(sdk.Identity.UserID, &apistructs.ScopeRoleAccessRequest{
		Scope: apistructs.Scope{
			Type: apistructs.ProjectScope,
			ID:   projectID,
		},
	})
	if err != nil {
		return false, err
	}
	for _, role := range scopeRole.Roles {
		if role == "Guest" {
			isGuest = true
		}
	}
	return isGuest, nil
}

func init() {
	base.InitProviderWithCreator("issue-manage", "issueImport", func() servicehub.Provider {
		return &ComponentAction{}
	})
}
