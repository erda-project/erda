// Copyright (c) 2021 Terminus, Inc.
//
// This program is free software: you can use, redistribute, and/or modify
// it under the terms of the GNU Affero General Public License, version 3
// or later ("AGPL"), as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

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
