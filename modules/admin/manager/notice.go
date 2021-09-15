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
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/admin/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (am *AdminManager) AppendNoticeEndpoint() {
	am.endpoints = append(am.endpoints, []httpserver.Endpoint{
		{Path: "/api/notices", Method: http.MethodPost, Handler: am.CreateNotice},
		{Path: "/api/notices/{id}", Method: http.MethodPut, Handler: am.UpdateNotice},
		{Path: "/api/notices/{id}/actions/publish", Method: http.MethodPut, Handler: am.PublishNotice},
		{Path: "/api/notices/{id}/actions/unpublish", Method: http.MethodPut, Handler: am.UnpublishNotice},
		{Path: "/api/notices/{id}", Method: http.MethodDelete, Handler: am.DeleteNotice},
		{Path: "/api/notices", Method: http.MethodGet, Handler: am.ListNotice},
	}...)
}

func (am *AdminManager) CreateNotice(contenxt context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {
	userID := req.Header.Get("USER-ID")
	id := USERID(userID)
	if id.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	orgID, err := GetOrgID(req)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	// check permission
	checkResp, err := am.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID,
		Scope:    apistructs.OrgScope,
		ScopeID:  orgID,
		Resource: apistructs.NoticeResource,
		Action:   apistructs.CreateAction,
	})
	if err != nil {
		return nil, err
	}
	if !checkResp.Access {
		return nil, apierrors.ErrCreateNotice.AccessDenied()
	}

	resp, err := am.bundle.CreateNoticeRequest(userID, orgID, req.Body)
	if err != nil {
		return apierrors.ErrCreateNotice.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(resp.Data)
}

func (am *AdminManager) UpdateNotice(contenxt context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {
	userID := req.Header.Get("USER-ID")
	uid := USERID(userID)
	if uid.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	orgID, err := GetOrgID(req)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	id, err := strconv.ParseUint(resources["id"], 10, 64)
	if err != nil {
		return apierrors.ErrUpdateNotice.InvalidParameter(err).ToResp(), nil
	}

	// check permission
	checkResp, err := am.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID,
		Scope:    apistructs.OrgScope,
		ScopeID:  orgID,
		Resource: apistructs.NoticeResource,
		Action:   apistructs.UpdateAction,
	})
	if err != nil {
		return nil, err
	}
	if !checkResp.Access {
		return nil, apierrors.ErrUpdateNotice.AccessDenied()
	}

	resp, err := am.bundle.UpdateNotice(id, orgID, userID, req.Body)
	if err != nil {
		return apierrors.ErrUpdateNotice.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(resp.Data)
}

func (am *AdminManager) PublishNotice(contenxt context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {
	userID := req.Header.Get("USER-ID")
	uid := USERID(userID)
	if uid.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	orgID, err := GetOrgID(req)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	id, err := strconv.ParseUint(resources["id"], 10, 64)
	if err != nil {
		return apierrors.ErrPublishNotice.InvalidParameter(err).ToResp(), nil
	}

	// check permission
	checkResp, err := am.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID,
		Scope:    apistructs.OrgScope,
		ScopeID:  orgID,
		Resource: apistructs.NoticeResource,
		Action:   apistructs.UpdateAction,
	})
	if err != nil {
		return nil, err
	}
	if !checkResp.Access {
		return nil, apierrors.ErrPublishNotice.AccessDenied()
	}

	err = am.bundle.PublishORUnPublishNotice(orgID, id, userID, "publish")
	if err != nil {
		return apierrors.ErrPublishNotice.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(nil)
}

func (am *AdminManager) UnpublishNotice(contenxt context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {
	userID := req.Header.Get("USER-ID")
	uid := USERID(userID)
	if uid.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	orgID, err := GetOrgID(req)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	id, err := strconv.ParseUint(resources["id"], 10, 64)
	if err != nil {
		return apierrors.ErrUnpublishNotice.InvalidParameter(err).ToResp(), nil
	}

	// check permission
	checkResp, err := am.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID,
		Scope:    apistructs.OrgScope,
		ScopeID:  orgID,
		Resource: apistructs.NoticeResource,
		Action:   apistructs.UpdateAction,
	})
	if err != nil {
		return nil, err
	}
	if !checkResp.Access {
		return nil, apierrors.ErrUnpublishNotice.AccessDenied()
	}

	err = am.bundle.PublishORUnPublishNotice(orgID, id, userID, "unpublish")
	if err != nil {
		return apierrors.ErrUnpublishNotice.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(nil)
}

func (am *AdminManager) DeleteNotice(contenxt context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {
	userID := req.Header.Get("USER-ID")
	uid := USERID(userID)
	if uid.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	orgID, err := GetOrgID(req)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	id, err := strconv.ParseUint(resources["id"], 10, 64)
	if err != nil {
		return apierrors.ErrDeleteNotice.InvalidParameter(err).ToResp(), nil
	}

	// check permission
	checkResp, err := am.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID,
		Scope:    apistructs.OrgScope,
		ScopeID:  orgID,
		Resource: apistructs.NoticeResource,
		Action:   apistructs.DeleteAction,
	})
	if err != nil {
		return nil, err
	}
	if !checkResp.Access {
		return nil, apierrors.ErrDeleteNotice.AccessDenied()
	}

	resp, err := am.bundle.DeleteNotice(id, orgID, userID)
	if err != nil {
		return apierrors.ErrDeleteNotice.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(resp.Data)
}

func (am *AdminManager) ListNotice(contenxt context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {

	userID := req.Header.Get("USER-ID")
	uid := USERID(userID)
	if uid.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	orgID, err := GetOrgID(req)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	// check permission
	checkResp, err := am.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID,
		Scope:    apistructs.OrgScope,
		ScopeID:  orgID,
		Resource: apistructs.NoticeResource,
		Action:   apistructs.ListAction,
	})
	if err != nil {
		return nil, err
	}
	if !checkResp.Access {
		return nil, apierrors.ErrListNotice.AccessDenied()
	}

	resp, err := am.bundle.ListNoticeByOrgID(orgID, userID, req.URL.Query())
	if err != nil {
		return apierrors.ErrListNotice.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(resp.Data, resp.UserIDs)
}
