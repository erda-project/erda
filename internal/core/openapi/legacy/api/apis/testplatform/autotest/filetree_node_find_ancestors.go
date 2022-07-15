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

package autotest

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/apis"
)

var FILETREE_NODE_FIND_ANCESTORS = apis.ApiSpec{
	Path:         "/api/autotests/filetree/<inode>/actions/find-ancestors",
	BackendPath:  "/api/autotests/filetree/<inode>/actions/find-ancestors",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       http.MethodGet,
	CheckLogin:   true,
	RequestType:  apistructs.UnifiedFileTreeNodeFindAncestorsRequest{},
	ResponseType: apistructs.UnifiedFileTreeNodeFindAncestorsResponse{},
	Doc:          "自动化测试目录树节点寻祖",
}
