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

var CMDB_ORG_GET_BY_DOMAIN = apis.ApiSpec{
	Path:          "/api/orgs/actions/get-by-domain",
	BackendPath:   "/api/orgs/actions/get-by-domain",
	Host:          "core-services.marathon.l4lb.thisdcos.directory:9526",
	Scheme:        "http",
	Method:        "GET",
	TryCheckLogin: true,
	CheckToken:    true,
	IsOpenAPI:     true,
	RequestType:   apistructs.OrgGetByDomainRequest{},
	ResponseType:  apistructs.OrgGetByDomainResponse{},
	Doc:           "summary: 通过域名获取组织",
}
