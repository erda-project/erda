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

package ops

import (
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var OPS_NODE_LABELS_UPDATE = apis.ApiSpec{
	Path:         "/api/node-labels",
	BackendPath:  "/api/node-labels",
	Host:         "ops.marathon.l4lb.thisdcos.directory:9027",
	Scheme:       "http",
	Method:       "POST",
	CheckLogin:   true,
	RequestType:  apistructs.UpdateLabelsRequest{},
	ResponseType: apistructs.UpdateLabelsResponse{},
	Doc:          "更新 node label",
	Audit: func(ctx *spec.AuditContext) error {
		var request apistructs.UpdateLabelsRequest
		if err := ctx.BindRequestData(&request); err != nil {
			return err
		}

		labels := strings.Join(request.Labels, ",")
		nodes := strings.Join(request.Hosts, ",")

		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.OrgScope,
			ScopeID:      uint64(ctx.OrgID),
			TemplateName: apistructs.UpdateNodeLabelsTemplate,
			Context: map[string]interface{}{
				"clusterName": request.ClusterName,
				"nodes":       nodes,
				"labels":      labels,
			},
		})
	},
}
