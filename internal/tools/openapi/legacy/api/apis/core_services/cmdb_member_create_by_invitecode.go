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

package core_services

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/tools/openapi/legacy/api/apis"
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
