package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

	if err := json.NewDecoder(req.Body).Decode(&listReq); err != nil {
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
	if int(listReq.OrgID) != 0 {
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
	ctx context.Context, writer http.ResponseWriter,
	req *http.Request, resources map[string]string) error {

	orgID, err := GetOrgID(req)
	if err != nil {
		return errors.Errorf("invalid param, orgId is invalid")
	}

	userID := req.Header.Get("USER-ID")
	id := USERID(userID)
	if id.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id"))
	}

	var listReq apistructs.AuditsListRequest
	if err := json.NewDecoder(req.Body).Decode(&listReq); err != nil {
		return apierrors.ErrListAudit.MissingParameter("body")
	}

	listReq.PageNo = 1
	listReq.PageSize = 99999
	if int(listReq.OrgID) != 0 {
		listReq.OrgID = orgID
	}

	// check params validate
	if err := listReq.Check(); err != nil {
		return apierrors.ErrListAudit.InvalidParameter(err)
	}

	err = am.bundle.ExportAuditExcel(&listReq, userID)
	if err != nil {
		return apierrors.ErrListAudit.InternalError(err)
	}
	return nil
}
