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
	"github.com/erda-project/erda/internal/tools/openapi/legacy/api/apis"
)

var CMDB_MANUAL_REVIEW_LISTBYSPONSORID = apis.ApiSpec{
	Path:         "/api/reviews/actions/list-launched-approval",
	BackendPath:  "/api/reviews/actions/list-launched-approval",
	Host:         "erda-server.marathon.l4lb.thisdcos.directory:9095",
	Scheme:       "http",
	Method:       "GET",
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	RequestType:  apistructs.GetReviewsBySponsorIdRequest{},
	ResponseType: apistructs.GetReviewsBySponsorIdResponse{},
	Doc:          "summary: 查询sponsorId(发起人)发起部署审批列表",
}
