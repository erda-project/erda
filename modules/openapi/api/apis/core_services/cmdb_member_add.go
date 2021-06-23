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
	"github.com/erda-project/erda/pkg/strutil"
)

var CMDB_MEMBER_ADD = apis.ApiSpec{
	Path:         "/api/members",
	BackendPath:  "/api/members",
	Host:         "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:       "http",
	Method:       "POST",
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	RequestType:  apistructs.MemberAddRequest{},
	ResponseType: apistructs.MemberAddResponse{},
	Doc:          "summary: 添加成员",
	Audit: func(ctx *spec.AuditContext) error {
		var requestBody apistructs.MemberAddRequest
		if err := ctx.BindRequestData(&requestBody); err != nil {
			return err
		}

		scopeID, err := strutil.Atoi64(requestBody.Scope.ID)
		if err != nil {
			return err
		}
		scopeType := requestBody.Scope.Type

		user, err := ctx.Bundle.ListUsers(apistructs.UserListRequest{UserIDs: requestBody.UserIDs})
		if err != nil {
			return err
		}

		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    scopeType,
			ScopeID:      uint64(scopeID),
			TemplateName: apistructs.AddMemberTemplate,
			Context:      map[string]interface{}{"users": user.Users},
		})
	},
}
