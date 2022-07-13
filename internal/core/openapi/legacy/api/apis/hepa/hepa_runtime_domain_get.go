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

package hepa

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/apis"
)

var HEPA_RUNTIME_DOMAIN_GET = apis.ApiSpec{
	Path:         "/api/runtimes/<runtimeID>/k8s-domains",
	BackendPath:  "/api/gateway/openapi/runtimes/<runtimeID>/domains",
	Host:         "hepa.marathon.l4lb.thisdcos.directory:8080",
	K8SHost:      "hepa.default.svc.cluster.local:8080",
	Scheme:       "http",
	Method:       "GET",
	CheckLogin:   true,
	RequestType:  apistructs.DomainListRequest{},
	ResponseType: apistructs.DomainListResponse{},
	Doc:          `获取 Runtime 域名`,
}
