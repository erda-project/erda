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

package cmdb

import (
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_CLUSTER_DEREFERENCE = apis.ApiSpec{
	Path:         "/api/clusters/actions/dereference",
	BackendPath:  "/api/clusters/actions/dereference",
	Host:         "cmdb.marathon.l4lb.thisdcos.directory:9093",
	Scheme:       "http",
	Method:       "PUT",
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	RequestType:  apistructs.DereferenceClusterRequest{},
	ResponseType: apistructs.DereferenceClusterResponse{},
	Doc:          "summary: 解除企业关联集群关系",
	Audit: func(ctx *spec.AuditContext) error {
		orgIDStr := ctx.Request.URL.Query().Get("orgID")
		if orgIDStr == "" {
			err := fmt.Errorf("get orgID failed")
			return err
		}

		clusterName := ctx.Request.URL.Query().Get("clusterName")
		if clusterName == "" {
			err := fmt.Errorf("get clusterName failed")
			return err
		}

		orgID, err := strconv.Atoi(orgIDStr)
		if err != nil {
			return err
		}

		return ctx.CreateAudit(&apistructs.Audit{
			OrgID:        uint64(orgID),
			ScopeType:    apistructs.OrgScope,
			ScopeID:      uint64(orgID),
			TemplateName: apistructs.ClusterDereferenceTemplate,
			Context: map[string]interface{}{
				"clusterName": clusterName,
			},
		})
	},
}
