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
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
	"github.com/erda-project/erda/modules/openapi/api/spec"
)

var OPS_CLOUD_RESOURCE_ONS_TOPIC_CREATE = apis.ApiSpec{
	Path:         "/api/cloud-ons/actions/create-topic",
	BackendPath:  "/api/cloud-ons/actions/create-topic",
	Host:         "ops.marathon.l4lb.thisdcos.directory:9027",
	Scheme:       "http",
	Method:       "POST",
	CheckLogin:   true,
	RequestType:  apistructs.CreateCloudResourceOnsTopicRequest{},
	ResponseType: apistructs.CreateCloudResourceOnsTopicResponse{},
	Doc:          "创建 ons topic",
	Audit: func(ctx *spec.AuditContext) error {
		var request apistructs.CreateCloudResourceOnsTopicRequest
		if err := ctx.BindRequestData(&request); err != nil {
			return err
		}

		if request.Vendor == "" || request.Vendor == "aliyun" {
			request.Vendor = "alicloud"
		}

		var topicList []string
		for _, t := range request.Topics {
			topicList = append(topicList, t.TopicName)
		}

		return ctx.CreateAudit(&apistructs.Audit{
			ScopeType:    apistructs.OrgScope,
			ScopeID:      uint64(ctx.OrgID),
			TemplateName: apistructs.CreateOnsTopicTemplate,
			Context: map[string]interface{}{
				"vendor":     request.Vendor,
				"instanceID": request.InstanceID,
				"topic":      strings.Join(topicList, ","),
			},
		})
	},
}
