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
)

var CMDB_MEMBER_ADD_BY_INVITECODE = apis.ApiSpec{
	Path:          "/api/members/actions/create-by-invitecode",
	BackendPath:   "/api/members/actions/create-by-invitecode",
	Host:          "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:        "http",
	Method:        "POST",
	TryCheckLogin: true,
	CheckToken:    true,
	IsOpenAPI:     true,
	RequestType:   apistructs.MemberAddByInviteCodeRequest{},
	ResponseType:  apistructs.MemberAddByInviteCodeResponse{},
	Doc:           "summary: 通过邀请码添加成员",
}
