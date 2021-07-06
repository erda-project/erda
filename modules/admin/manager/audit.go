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
	"io"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/admin/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (am *AdminManager) AppendAuditEndpoint() {
	am.endpoints = append(am.endpoints, []httpserver.Endpoint{
		{Path: "/api/audits/actions/list", Method: http.MethodGet, Handler: am.ListAudits},
		{Path: "/api/audits/actions/export-excel", Method: http.MethodGet, WriterHandler: am.ExportExcelAudit},
	}...)
}

func (am *AdminManager) ListAudits(ctx context.Context, req *http.Request, vars map[string]string) (httpserver.Responser, error) {

	var listReq apistructs.AuditsListRequest
	queryDecoder := schema.NewDecoder()
	queryDecoder.IgnoreUnknownKeys(true)

	if err := queryDecoder.Decode(&listReq, req.URL.Query()); err != nil {
		return apierrors.ErrListAudit.MissingParameter("body").ToResp(), nil
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

	if listReq.OrgID == 0 {
		listReq.OrgID = orgID
	}

	// check params validate
	if err := listReq.Check(); err != nil {
		return apierrors.ErrListAudit.InvalidParameter(err).ToResp(), nil
	}

	resp, err := am.bundle.ListAuditEvent(&listReq, userID)
	if err != nil {
		return apierrors.ErrListAudit.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(resp.Data, resp.UserIDs)
}

func (am *AdminManager) ExportExcelAudit(
	ctx context.Context, w http.ResponseWriter,
	req *http.Request, resources map[string]string) error {

	var listReq apistructs.AuditsListRequest
	queryDecoder := schema.NewDecoder()
	queryDecoder.IgnoreUnknownKeys(true)

	if err := queryDecoder.Decode(&listReq, req.URL.Query()); err != nil {
		return apierrors.ErrListAudit.MissingParameter("body")
	}

	orgID, err := GetOrgID(req)
	if err != nil {
		return errors.Errorf("invalid param, orgId is invalid")
	}

	userID := req.Header.Get("USER-ID")
	id := USERID(userID)
	if id.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id"))
	}

	listReq.PageNo = 1
	listReq.PageSize = 99999
	if listReq.OrgID == 0 {
		listReq.OrgID = orgID
	}

	// check params validate
	if err := listReq.Check(); err != nil {
		return apierrors.ErrListAudit.InvalidParameter(err)
	}

	respBody, resp, err := am.bundle.ExportAuditExcel(&listReq, userID)
	if err != nil {
		return fmt.Errorf("failed to get spec from file: %v", err)
	}
	w.Header().Set("Content-Disposition", resp.Headers().Get("Content-Disposition"))
	w.Header().Set("Content-Type", resp.Headers().Get("Content-Type"))
	_, err = io.Copy(w, respBody)
	return err
}
