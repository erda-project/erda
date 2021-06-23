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

package core_services

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_ORG_CLUSTER_RELATE_OPENAPI = apis.ApiSpec{
	Path:        "/api/orgs/actions/relate-cluster",
	BackendPath: "/api/orgs/actions/relate-cluster",
	Host:        "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:      "http",
	Method:      "POST",
	IsOpenAPI:   true,
	RequestType: apistructs.OrgClusterRelationCreateRequest{},
	CheckLogin:  true,
	CheckToken:  true,
	Doc:         "summary: 添加企业集群关联",
	Audit: func(ctx *spec.AuditContext) error {
		var request apistructs.OrgClusterRelationCreateRequest
		if err := ctx.BindRequestData(&request); err != nil {
			return err
		}

		return ctx.CreateAudit(&apistructs.Audit{
			OrgID:        request.OrgID,
			ScopeType:    apistructs.OrgScope,
			ScopeID:      request.OrgID,
			TemplateName: apistructs.ClusterReferenceTemplate,
			Context: map[string]interface{}{
				"clusterName": request.ClusterName,
			},
		})
	},
}
