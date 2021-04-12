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

package gittar

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var REPO_POST = apis.ApiSpec{
	Path:        "/api/repo/<*>",
	BackendPath: "/wb/<*>",
	Host:        "gittar.marathon.l4lb.thisdcos.directory:5566",
	Scheme:      "http",
	Method:      "POST",
	CheckLogin:  true,
	IsOpenAPI:   true,
	CheckToken:  true,
	Doc:         `summary: repo delete api proxy`,
	Audit: func(ctx *spec.AuditContext) error {
		var responseBody apistructs.LockedRepoResponse
		if err := ctx.BindResponseData(&responseBody); err != nil {
			return err
		}
		if responseBody.Data.AppName == "" {
			return nil
		}
		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.AppScope,
			ScopeID:      1,
			OrgID:        uint64(ctx.OrgID),
			TemplateName: apistructs.RepoLockedTemplate,
			AppID:        uint64(responseBody.Data.AppID),
			ProjectID:    uint64(responseBody.Data.ProjectID),
			Context: map[string]interface{}{
				"appName":   responseBody.Data.AppName,
				"isLocked":  strconv.FormatBool(responseBody.Data.IsLocked),
				"appId":     responseBody.Data.AppID,
				"projectId": responseBody.Data.ProjectID,
			},
		})
	},
}
