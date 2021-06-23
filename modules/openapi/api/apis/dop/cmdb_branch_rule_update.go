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

package dop

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_BRANCH_RULE_UPDATE = apis.ApiSpec{
	Path:         "/api/branch-rules/<id>",
	BackendPath:  "/api/branch-rules/<id>",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       "PUT",
	CheckLogin:   true,
	RequestType:  apistructs.UpdateBranchRuleRequest{},
	ResponseType: apistructs.UpdateBranchRuleResponse{},
	Doc:          "summary: 更新分支规则",
	Audit: func(ctx *spec.AuditContext) error {
		var resp apistructs.UpdateBranchRuleResponse
		err := ctx.BindResponseData(&resp)
		if err != nil {
			return err
		}
		ruleDTO := resp.Data
		if ruleDTO.ScopeType != apistructs.ProjectScope {
			return nil
		}
		project, err := ctx.GetProject(ruleDTO.ScopeID)
		if err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			Context: map[string]interface{}{
				"projectId":   strconv.FormatUint(project.ID, 10),
				"projectName": project.Name,
				"ruleName":    ruleDTO.Rule,
			},
			ScopeType:    "project",
			ScopeID:      project.ID,
			Result:       apistructs.SuccessfulResult,
			TemplateName: apistructs.UpdateBranchRuleTemplate,
		})
	},
}
