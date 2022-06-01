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

package gittar

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/tools/openapi/legacy/api/apis"
	"github.com/erda-project/erda/modules/tools/openapi/legacy/api/spec"
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
