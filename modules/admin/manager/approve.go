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

package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

func (am *AdminManager) AppendApproveEndpoint() {
	am.endpoints = append(am.endpoints, []httpserver.Endpoint{
		{Path: "/api/approves/actions/list-approves", Method: http.MethodGet, Handler: am.ListApprove},
		{Path: "/api/approves/{approveId}", Method: http.MethodGet, Handler: am.GetApprove},
		{Path: "/api/approves/{approveId}", Method: http.MethodPut, Handler: am.UpdateApprove},
	}...)
}

func (am *AdminManager) ListApprove(contenxt context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {

	listReq, err := getListApprovesParam(req)
	if err != nil {
		return apierrors.ErrListApprove.InvalidParameter(err).ToResp(), nil
	}

	userID := req.Header.Get("USER-ID")
	id := USERID(userID)
	if id.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	resp, err := am.bundle.ListApprove(listReq, userID)
	if err != nil {
		return apierrors.ErrListApprove.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(resp.Data, resp.UserIDs)
}

func (am *AdminManager) GetApprove(contenxt context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {

	approveID, err := strutil.Atoi64(resources["approveId"])
	if err != nil {
		return apierrors.ErrGetApprove.InvalidParameter(err).ToResp(), nil
	}

	userID := req.Header.Get("USER-ID")
	id := USERID(userID)
	if id.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	orgID, err := GetOrgID(req)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	resp, err := am.bundle.GetApprove(fmt.Sprintf("%d", orgID), userID, approveID)
	if err != nil {
		return apierrors.ErrGetApprove.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(resp.Data)
}

func (am *AdminManager) UpdateApprove(contenxt context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {

	approveID, err := strutil.Atoi64(resources["approveId"])
	if err != nil {
		return apierrors.ErrGetApprove.InvalidParameter(err).ToResp(), nil
	}

	// validate approveID
	if approveID == 0 {
		return apierrors.ErrUpdateApprove.InvalidParameter("need approveId").ToResp(), nil
	}

	userID := req.Header.Get("USER-ID")
	id := USERID(userID)
	if id.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	orgID, err := GetOrgID(req)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	var approveUpdateReq apistructs.ApproveUpdateRequest
	if req.Body == nil {
		return apierrors.ErrUpdateApprove.MissingParameter("body").ToResp(), nil
	}
	if err := json.NewDecoder(req.Body).Decode(&approveUpdateReq); err != nil {
		return apierrors.ErrUpdateApprove.InvalidParameter(err).ToResp(), nil
	}
	approveUpdateReq.OrgID = orgID

	resp, err := am.bundle.UpdateApprove(approveUpdateReq, userID, approveID)
	if err != nil {
		return apierrors.ErrUpdateApprove.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(resp.Data)
}

// Approve列表时获取请求参数
func getListApprovesParam(r *http.Request) (*apistructs.ApproveListRequest, error) {
	orgID, err := GetOrgID(r)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	var status []string
	statusMap := r.URL.Query()
	if statusList, ok := statusMap["status"]; ok {
		for _, s := range statusList {
			if s != string(apistructs.ApprovalStatusPending) &&
				s != string(apistructs.ApprovalStatusApproved) &&
				s != string(apistructs.ApprovalStatusDeined) {
				return nil, errors.Errorf("status type error")
			}
			status = append(status, s)
		}
	}

	// 获取pageSize
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return nil, errors.Errorf("invalid param, pageSize is invalid")
	}
	// 获取pageNo
	pageNoStr := r.URL.Query().Get("pageNo")
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	pageNo, err := strconv.Atoi(pageNoStr)
	if err != nil {
		return nil, errors.Errorf("invalid param, pageNo is invalid")
	}
	var id *int64
	id_str := r.URL.Query().Get("id")
	if id_str != "" {
		id_int, err := strconv.ParseInt(id_str, 10, 64)
		if err != nil {
			return nil, errors.Errorf("invalid param, id is invalid")
		}
		id = &id_int
	}

	return &apistructs.ApproveListRequest{
		OrgID:    orgID,
		Status:   status,
		PageNo:   pageNo,
		PageSize: pageSize,
		ID:       id,
	}, nil
}
