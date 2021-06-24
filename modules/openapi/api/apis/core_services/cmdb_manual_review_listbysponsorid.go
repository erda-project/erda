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

var CMDB_MANUAL_REVIEW_LISTBYSPONSORID = apis.ApiSpec{
	Path:         "/api/reviews/actions/list-launched-approval",
	BackendPath:  "/api/reviews/actions/list-launched-approval",
	Host:         "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:       "http",
	Method:       "GET",
	CheckLogin:   true,
	CheckToken:   true,
	IsOpenAPI:    true,
	RequestType:  apistructs.GetReviewsBySponsorIdRequest{},
	ResponseType: apistructs.GetReviewsBySponsorIdResponse{},
	Doc:          "summary: 查询sponsorId(发起人)发起部署审批列表",
}
