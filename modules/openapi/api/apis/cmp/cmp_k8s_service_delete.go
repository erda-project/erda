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
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var CMP_K8S_SERVICE_DELETE = apis.ApiSpec{
	Path:        "/apis/clusters/<clusterName>/namespaces/<namespaceName>/services/<serviceName>",
	BackendPath: "/apis/clusters/<clusterName>/namespaces/<namespaceName>/services/<serviceName>",
	Host:        "cmp.marathon.l4lb.thisdcos.directory:9027",
	Scheme:      "http",
	Method:      "DELETE",
	CheckLogin:  true,
	Doc:         "删除 k8s service 对象",
}
