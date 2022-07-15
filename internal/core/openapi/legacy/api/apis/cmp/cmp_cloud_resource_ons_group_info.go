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

package cmp

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/apis"
)

var CMP_CLOUD_RESOURCE_ONS_GROUP_INFO = apis.ApiSpec{
	Path:         "/api/cloud-ons/actions/list-group",
	BackendPath:  "/api/cloud-ons/actions/list-group",
	Host:         "cmp.marathon.l4lb.thisdcos.directory:9027",
	Scheme:       "http",
	Method:       "GET",
	CheckLogin:   true,
	RequestType:  apistructs.CloudResourceOnsGroupInfoRequest{},
	ResponseType: apistructs.CloudResourceOnsGroupInfoResponse{},
	Doc:          "query参数传入topicName时，返回相应topic信息；未传入则返回所有topic信息",
}
