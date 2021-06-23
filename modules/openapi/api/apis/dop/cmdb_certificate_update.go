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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var CMDB_CERTIFICATE_UPDATE = apis.ApiSpec{
	Path:         "/api/certificates/<certificateID>",
	BackendPath:  "/api/certificates/<certificateID>",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       "PUT",
	CheckLogin:   true,
	RequestType:  apistructs.CertificateUpdateRequest{},
	ResponseType: apistructs.CertificateUpdateResponse{},
	Doc:          "summary: 更新证书",
	Audit: func(ctx *spec.AuditContext) error {

		var resp apistructs.CertificateUpdateResponse
		err := ctx.BindResponseData(&resp)
		if err != nil {
			return err
		}
		if resp.Success {
			return ctx.CreateAudit(&apistructs.Audit{
				ScopeType:    apistructs.OrgScope,
				ScopeID:      uint64(ctx.OrgID),
				TemplateName: apistructs.UpdateCertificatesTemplate,
				Context:      map[string]interface{}{"certificateName": resp.Data.Name},
			})

		} else {
			return nil
		}
	},
}
