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
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (b *Bundle) GetWorkbenchData(userID string, req apistructs.WorkbenchRequest) (*apistructs.WorkbenchResponse, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var rsp apistructs.WorkbenchResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/workbench/actions/list")).
		Header(httputil.UserHeader, userID).
		Param("pageNo", strconv.FormatInt(int64(req.PageNo), 10)).
		Param("pageSize", strconv.FormatInt(int64(req.PageSize), 10)).
		Param("issueSize", strconv.FormatInt(int64(req.IssueSize), 10)).
		Param("orgID", strconv.FormatInt(int64(req.OrgID), 10)).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrGetWorkBenchData.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return &rsp, nil
}

func (b *Bundle) GetIssuesForWorkbench(req apistructs.IssuePagingRequest) (*apistructs.IssuePagingResponse, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var rsp apistructs.IssuePagingResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/workbench/issues/actions/list")).
		Header(httputil.UserHeader, req.UserID).
		Params(req.UrlQueryString()).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrGetWorkBenchData.InternalError(err)
	}
	if !httpResp.IsOK() || !rsp.Success {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}

	return &rsp, nil
}
