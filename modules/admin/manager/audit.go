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
	userID := req.Header.Get("USER-ID")
	id := USERID(userID)
	if id.Invalid() {
		return apierrors.ErrListAudit.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	var orgIDStr = ""
	if req.URL.Query().Get("sys") == "" {
		id, err := GetOrgID(req)
		if err != nil {
			return apierrors.ErrListAudit.InvalidParameter(err).ToResp(), nil
		}
		orgIDStr = fmt.Sprintf("%d", id)
	}

	resp, err := am.bundle.ListAuditEvent(orgIDStr, userID, req.URL.Query())
	if err != nil {
		return apierrors.ErrListAudit.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(resp.Data, resp.UserIDs)
}

func (am *AdminManager) ExportExcelAudit(
	ctx context.Context, w http.ResponseWriter,
	req *http.Request, resources map[string]string) error {

	userID := req.Header.Get("USER-ID")
	id := USERID(userID)
	if id.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id"))
	}

	var orgIDStr = ""
	if req.URL.Query().Get("sys") == "" {
		id, err := GetOrgID(req)
		if err != nil {
			return apierrors.ErrListAudit.InvalidParameter(err)
		}
		orgIDStr = fmt.Sprintf("%d", id)
	}

	respBody, resp, err := am.bundle.ExportAuditExcel(orgIDStr, userID, req.URL.Query())
	if err != nil {
		return fmt.Errorf("failed to get spec from file: %v", err)
	}
	w.Header().Set("Content-Disposition", resp.Headers().Get("Content-Disposition"))
	w.Header().Set("Content-Type", resp.Headers().Get("Content-Type"))
	_, err = io.Copy(w, respBody)

	return err
}
