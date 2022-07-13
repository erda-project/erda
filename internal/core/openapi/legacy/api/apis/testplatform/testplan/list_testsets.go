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

package testplan

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/openapi/legacy/api/apis"
)

var LIST_TESTSETS = apis.ApiSpec{
	Path:         "/api/testplans/<testPlanID>/testsets",
	BackendPath:  "/api/testplans/<testPlanID>/testsets",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       http.MethodGet,
	CheckLogin:   true,
	RequestType:  apistructs.TestPlanTestSetsListRequest{},
	ResponseType: apistructs.TestPlanTestSetListResponse{},
	Doc:          "summary: 获取测试计划下的测试集列表",
}
