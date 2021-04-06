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

package bundle

import (
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

// ExecuteApiTest 执行接口测试
func (b *Bundle) ExecuteApiTest(req apistructs.ApiTestsActionRequest) (uint64, error) {
	host, err := b.urls.QA()
	if err != nil {
		return 0, err
	}
	hc := b.hc

	var executeResp apistructs.ApiTestsActionResponse
	resp, err := hc.Post(host).Path("/api/apitests/actions/execute-tests").
		Header(httputil.InternalHeader, "bundle").
		JSONBody(&req).
		Do().JSON(&executeResp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !executeResp.Success {
		return 0, toAPIError(resp.StatusCode(), executeResp.Error)
	}

	return executeResp.Data, nil
}

// InternalRemoveTestPlanCaseRelIssueRelationsByIssueID 内部使用，根据 issueID 删除测试计划用例与 bug 的关联关系
func (b *Bundle) InternalRemoveTestPlanCaseRelIssueRelationsByIssueID(issueID uint64) error {
	host, err := b.urls.QA()
	if err != nil {
		return err
	}
	hc := b.hc

	var removeResp struct {
		apistructs.Header
	}
	resp, err := hc.Delete(host).Path("/api/testplans/testcase-relations/actions/internal-remove-issue-relations").
		Header(httputil.InternalHeader, "bundle").
		Param("issueID", strconv.FormatUint(issueID, 10)).
		Do().JSON(&removeResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !removeResp.Success {
		return toAPIError(resp.StatusCode(), removeResp.Error)
	}

	return nil
}
