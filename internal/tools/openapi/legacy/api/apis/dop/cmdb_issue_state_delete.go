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

package dop

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/tools/openapi/legacy/api/apis"
	"github.com/erda-project/erda/internal/tools/openapi/legacy/api/spec"
)

var CMDB_ISSUE_STATE_DELETE = apis.ApiSpec{
	Path:         "/api/issues/actions/delete-state",
	BackendPath:  "/api/issues/actions/delete-state",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       http.MethodDelete,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  apistructs.IssueStateDeleteRequest{},
	ResponseType: apistructs.IssueStateDeleteResponse{},
	IsOpenAPI:    true,
	Doc:          "summary: 删除 issueState",
	Audit: func(ctx *spec.AuditContext) error {
		var responseBody apistructs.IssueStateDeleteResponse
		if err := ctx.BindResponseData(&responseBody); err != nil {
			return err
		}
		projectID := uint64(responseBody.Data.ProjectID)
		project, err := ctx.Bundle.GetProject(projectID)
		if err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.ProjectScope,
			ScopeID:      projectID,
			ProjectID:    projectID,
			TemplateName: apistructs.DeleteIssueStateTemplate,
			Context: map[string]interface{}{
				"projectName": project.Name,
				"issueType":   responseBody.Data.IssueType,
				"stateName":   responseBody.Data.StateName,
			},
		})
	},
}
