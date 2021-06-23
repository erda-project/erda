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

var CMDB_CERTIFICATE_DELETE = apis.ApiSpec{
	Path:         "/api/certificates/<certificatesID>",
	BackendPath:  "/api/certificates/<certificatesID>",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       "DELETE",
	CheckLogin:   true,
	ResponseType: apistructs.CertificateDeleteResponse{},
	Doc:          "summary: 删除证书",
	Audit: func(ctx *spec.AuditContext) error {

		var resp apistructs.CertificateDeleteResponse
		err := ctx.BindResponseData(&resp)
		if err != nil {
			return err
		}
		if resp.Success {
			return ctx.CreateAudit(&apistructs.Audit{
				ScopeType:    apistructs.OrgScope,
				ScopeID:      uint64(ctx.OrgID),
				TemplateName: apistructs.DeleteCertificatesTemplate,
				Context:      map[string]interface{}{"certificateName": resp.Data.Name},
			})

		} else {
			return nil
		}

	},
}
