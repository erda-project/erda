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
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// GetIteration 通过id获取迭代
func (b *Bundle) GetIteration(id uint64) (*apistructs.Iteration, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var iterationResp apistructs.IterationGetResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/iterations/%d", id)).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&iterationResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !iterationResp.Success {
		return nil, toAPIError(resp.StatusCode(), iterationResp.Error)
	}

	return &iterationResp.Data, nil
}

// ListProjectIterations 查询项目迭代
func (b *Bundle) ListProjectIterations(req apistructs.IterationPagingRequest, orgID string) ([]apistructs.Iteration, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var listResp apistructs.IterationPagingResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/iterations")).
		Param("projectID", strconv.FormatUint(req.ProjectID, 10)).
		Param("pageNo", strconv.FormatUint(req.PageNo, 10)).
		Param("pageSize", strconv.FormatUint(req.PageSize, 10)).
		Param("withoutIssueSummary", strconv.FormatBool(req.WithoutIssueSummary)).
		Header(httputil.OrgHeader, orgID).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&listResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !listResp.Success {
		return nil, toAPIError(resp.StatusCode(), listResp.Error)
	}

	return listResp.Data.List, nil
}
