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
	"net/http"

	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var APIDocWebsocket = apis.ApiSpec{
	Path:         "/api/k8s/clusters/<*>",
	BackendPath:  "/api/k8s/clusters/<*>",
	Host:         "cmp.marathon.l4lb.thisdcos.directory:9027",
	Scheme:       "ws",
	Method:       http.MethodGet,
	CheckLogin:   true,
	CheckToken:   true,
	RequestType:  nil,
	ResponseType: nil,
	Doc:          "k8s shell and logs",
}
