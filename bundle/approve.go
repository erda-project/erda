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

func (b *Bundle) CreateApprove(req *apistructs.ApproveCreateRequest, userID string) (*apistructs.ApproveDTO, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var createResp apistructs.ApproveCreateResponse
	resp, err := hc.Post(host).Path("/api/approves").
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.UserHeader, userID).
		JSONBody(req).Do().JSON(&createResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !createResp.Success {
		return nil, toAPIError(resp.StatusCode(), createResp.Error)
	}

	return &createResp.Data, nil
}

func (b *Bundle) ListApprove(req *apistructs.ApproveListRequest, userID string) (*apistructs.ApproveListResponse, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}

	hc := b.hc

	var approveList apistructs.ApproveListResponse
	resp, err := hc.Get(host).Path("/api/approves/actions/list-approves").
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.OrgHeader, fmt.Sprintf("%d", req.OrgID)).
		Header(httputil.UserHeader, userID).
		Param("pageSize", strconv.Itoa(req.PageSize)).
		Param("pageNo", strconv.Itoa(req.PageNo)).
		Do().
		JSON(&approveList)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to list approve, status code: %d, body: %v",
				resp.StatusCode(),
				resp.Body(),
			))
	}
	return &approveList, nil
}

func (b *Bundle) GetApprove(orgID, userID string, approveID int64) (*apistructs.ApproveDetailResponse, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var approve apistructs.ApproveDetailResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/approves/%d", approveID)).
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.OrgHeader, orgID).
		Header(httputil.UserHeader, userID).
		Do().
		JSON(&approve)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to get approve, status code: %d, body: %v",
				resp.StatusCode(),
				string(resp.Body()),
			))
	}

	return &approve, nil
}

func (b *Bundle) UpdateApprove(approve apistructs.ApproveUpdateRequest, userID string, approveID int64) (
	*apistructs.ApproveUpdateResponse, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var updateApprove apistructs.ApproveUpdateResponse
	resp, err := hc.Put(host).Path(fmt.Sprintf("/api/approves/%d", approveID)).
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.OrgHeader, strconv.Itoa(int(approve.OrgID))).
		Header(httputil.UserHeader, userID).
		JSONBody(approve).
		Do().
		JSON(&updateApprove)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return nil, apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to update approve, status code: %d, body: %v",
				resp.StatusCode(),
				string(resp.Body()),
			))
	}
	return &updateApprove, nil
}
