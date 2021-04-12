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

package ops

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var OPS_CLOUD_RESOURCE_ONS_TOPIC_INFO = apis.ApiSpec{
	Path:         "/api/cloud-ons/actions/list-topic",
	BackendPath:  "/api/cloud-ons/actions/list-topic",
	Host:         "ops.marathon.l4lb.thisdcos.directory:9027",
	Scheme:       "http",
	Method:       "GET",
	CheckLogin:   true,
	RequestType:  apistructs.CloudResourceOnsTopicInfoRequest{},
	ResponseType: apistructs.CloudResourceOnsTopicInfoResponse{},
	Doc:          "query参数传入topicName时，返回相应topic信息；未传入则返回所有topic信息",
}
