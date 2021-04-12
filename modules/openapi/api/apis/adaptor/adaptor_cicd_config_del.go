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

package adaptor

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var ADAPTOR_CICD_CONFIG_DEL = apis.ApiSpec{
	Path:        "/api/cicds/configs",
	BackendPath: "/api/cicds/configs",
	Host:        "gittar-adaptor.marathon.l4lb.thisdcos.directory:1086",
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
