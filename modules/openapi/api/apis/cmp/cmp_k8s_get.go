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

package cmp

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var CMP_STEVE_GET = apis.ApiSpec{
	Path:         "/api/k8s/clusters/<*>",
	BackendPath:  "/api/k8s/clusters/<*>",
	Method:       "GET",
	Host:         "cmp.marathon.l4lb.thisdcos.directory:9027",
	K8SHost:      "cmp:9027",
	Scheme:       "http",
	Audit:        nil,
	CheckLogin:   true,
	Doc:          "get a k8s resource",
	ResponseType: apistructs.SteveResource{},
	IsOpenAPI:    true,
}
