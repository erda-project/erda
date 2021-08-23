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

	"github.com/erda-project/erda/apistructs"
	protocol "github.com/erda-project/erda/modules/openapi/component-protocol"
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

type ComponentAction struct{}

func (ca *ComponentAction) Render(ctx context.Context, c *apistructs.Component, scenario apistructs.ComponentProtocolScenario, event apistructs.ComponentEvent, gs *apistructs.GlobalStateData) error {
	ctxBdl := ctx.Value(protocol.GlobalInnerKeyCtxBundle.String()).(protocol.ContextBundle)
	c.Props = IssueImportProps{
		Size:       "small",
		Tooltip:    "导入",
		PrefixIcon: "import",
		Visible:    ctxBdl.InParams["fixedIssueType"].(string) != "ALL",
	}

	isGuest, err := ca.CheckUserPermission(ctxBdl)
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
func (ca *ComponentAction) CheckUserPermission(bdl protocol.ContextBundle) (bool, error) {
	isGuest := false
	projectID := bdl.InParams["projectId"].(string)
	scopeRole, err := bdl.Bdl.ScopeRoleAccess(bdl.Identity.UserID, &apistructs.ScopeRoleAccessRequest{
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

func RenderCreator() protocol.CompRender {
	return &ComponentAction{}
}
