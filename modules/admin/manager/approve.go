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
	"fmt"
	"net/http"

	"github.com/pkg/errors"

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

	userID := req.Header.Get("USER-ID")
	id := USERID(userID)
	if id.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	orgID, err := GetOrgID(req)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	resp, err := am.bundle.ListApprove(orgID, userID, req.URL.Query())
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

	if req.Body == nil {
		return apierrors.ErrUpdateApprove.MissingParameter("body").ToResp(), nil
	}

	resp, err := am.bundle.UpdateApprove(orgID, userID, approveID, req.Body)
	if err != nil {
		return apierrors.ErrUpdateApprove.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(resp.Data)
}
