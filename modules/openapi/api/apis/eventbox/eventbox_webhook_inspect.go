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

package eventbox

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var EVENTBOX_WEBHOOK_INSPECT = apis.ApiSpec{
	Path:         "/api/webhooks/<id>",
	BackendPath:  "/api/dice/eventbox/webhooks/<id>",
	Host:         "eventbox.marathon.l4lb.thisdcos.directory:9528",
	Scheme:       "http",
	Method:       "GET",
	RequestType:  apistructs.WebhookInspectRequest{},
	ResponseType: apistructs.WebhookInspectResponse{},
	CheckLogin:   true,
	Doc:          `获取 webhook 详情`,
	IsOpenAPI:    true,
}
