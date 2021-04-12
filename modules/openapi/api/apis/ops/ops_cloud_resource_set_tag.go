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

var OPS_CLOUD_RESOURCE_SET_TAG = apis.ApiSpec{
	Path:         "/api/cloud-resource/set-tag",
	BackendPath:  "/api/cloud-resource/set-tag",
	Host:         "ops.marathon.l4lb.thisdcos.directory:9027",
	Scheme:       "http",
	Method:       "POST",
	CheckLogin:   true,
	RequestType:  apistructs.CloudResourceSetTagRequest{},
	ResponseType: apistructs.CloudResourceSetTagResponse{},
	Doc:          "tag cluster on vpc",
	Audit: func(ctx *spec.AuditContext) error {
		var request apistructs.CloudResourceSetTagRequest
		if err := ctx.BindRequestData(&request); err != nil {
			return err
		}

		var rsList []string
		var vendor string
		for _, v := range request.Items {
			vendor = v.Vendor
			rsList = append(rsList, v.ResourceID)
		}

		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.OrgScope,
			ScopeID:      uint64(ctx.OrgID),
			TemplateName: apistructs.SetCRTagsTemplate,
			Context: map[string]interface{}{
				"vendor":       vendor,
				"resourceType": request.ResourceType,
				"instanceID":   request.InstanceID,
				"crItems":      strings.Join(rsList, ","),
				"labels":       strings.Join(request.Tags, ","),
			},
		})
	},
}
