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

package rel

import (
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/openapi/api/apis"
)

var ADD_ISSUE_RELATION = apis.ApiSpec{
	Path:         "/api/testplans/<testPlanID>/testcase-relations/<relationID>/actions/add-issue-relations",
	BackendPath:  "/api/testplans/<testPlanID>/testcase-relations/<relationID>/actions/add-issue-relations",
	Host:         "dop.marathon.l4lb.thisdcos.directory:9527",
	Scheme:       "http",
	Method:       http.MethodPost,
	CheckLogin:   true,
	RequestType:  apistructs.TestPlanCaseRelIssueRelationAddRequest{},
	ResponseType: apistructs.TestPlanCaseRelIssueRelationAddResponse{},
	Doc:          "summary: 新增测试计划用例里的缺陷关联关系",
}
