// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gorilla/schema"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/core/legacy/services/apierrors"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (am *AdminManager) AppendUserEndpoint() {
	am.endpoints = append(am.endpoints, []httpserver.Endpoint{
		{Path: "/api/users", Method: http.MethodGet, Handler: am.ListUser},
		{Path: "/api/users/actions/search", Method: http.MethodGet, Handler: am.SearchUser},
	}...)
}

func (am *AdminManager) ListUser(ctx context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {
	queryDecoder := schema.NewDecoder()
	queryDecoder.IgnoreUnknownKeys(true)

	listReq := apistructs.UserListRequest{}
	if err := queryDecoder.Decode(&listReq, req.URL.Query()); err != nil {
		return apierrors.ErrListUser.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(req)
	if err != nil {
		return apierrors.ErrListUser.NotLogin().ToResp(), nil
	}
	var orgID uint64
	orgIDStr := req.Header.Get(httputil.OrgHeader)
	if orgIDStr != "" {
		orgID, err = strconv.ParseUint(orgIDStr, 10, 64)
		if err != nil {
			return apierrors.ErrListUser.InvalidParameter("orgId is invalid").ToResp(), nil
		}
	}

	res, err := am.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   identityInfo.UserID,
		Scope:    apistructs.OrgScope,
		ScopeID:  orgID,
		Resource: apistructs.MemberResource,
		Action:   apistructs.CreateAction,
	})
	if err != nil {
		return apierrors.ErrListUser.InternalError(err).ToResp(), nil
	}
	if !res.Access {
		return apierrors.ErrListUser.AccessDenied().ToResp(), nil
	}

	resp, err := am.bundle.ListUsers(listReq)
	if err != nil {
		return apierrors.ErrListUser.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(resp)
}

func (am *AdminManager) SearchUser(ctx context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {
	queryDecoder := schema.NewDecoder()
	queryDecoder.IgnoreUnknownKeys(true)

	identityInfo, err := user.GetIdentityInfo(req)
	if err != nil {
		return apierrors.ErrListUser.NotLogin().ToResp(), nil
	}
	var orgID uint64
	orgIDStr := req.Header.Get(httputil.OrgHeader)
	if orgIDStr != "" {
		orgID, err = strconv.ParseUint(orgIDStr, 10, 64)
		if err != nil {
			return apierrors.ErrListUser.InvalidParameter("orgId is invalid").ToResp(), nil
		}
	}

	res, err := am.bundle.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   identityInfo.UserID,
		Scope:    apistructs.OrgScope,
		ScopeID:  orgID,
		Resource: apistructs.MemberResource,
		Action:   apistructs.CreateAction,
	})
	if err != nil {
		return apierrors.ErrListUser.InternalError(err).ToResp(), nil
	}
	if !res.Access {
		return apierrors.ErrListUser.AccessDenied().ToResp(), nil
	}

	resp, err := am.bundle.SearchUser(req.URL.Query())
	if err != nil {
		return apierrors.ErrListUser.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(resp)
}
