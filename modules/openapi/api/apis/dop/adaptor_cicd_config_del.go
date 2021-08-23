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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var ADAPTOR_CICD_CONFIG_DEL = apis.ApiSpec{
	Path:        "/api/cicds/configs",
	BackendPath: "/api/cicds/configs",
	Host:        "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:      "http",
	Method:      "DELETE",
	CheckLogin:  true,
	Doc:         "summary: 删除Pipeline指定命名空间下的某个配置",
	Audit: func(ctx *spec.AuditContext) error {
		appID := ctx.Request.URL.Query().Get("appID")
		namespaceName := ctx.Request.URL.Query().Get("namespace_name")
		key := ctx.Request.URL.Query().Get("key")
		appDTO, err := ctx.GetApp(appID)
		if err != nil {
			return err
		}
		return ctx.CreateAudit(&apistructs.Audit{
			Context: map[string]interface{}{
				"projectId":   strconv.FormatUint(appDTO.ProjectID, 10),
				"appId":       strconv.FormatUint(appDTO.ID, 10),
				"projectName": appDTO.ProjectName,
				"appName":     appDTO.Name,
				"namespace":   namespaceName,
				"key":         key,
			},
			ProjectID:    appDTO.ProjectID,
			AppID:        appDTO.ID,
			ScopeType:    "app",
			ScopeID:      appDTO.ID,
			TemplateName: apistructs.DeletePipelineKeyTemplate,
		})
	},
}
