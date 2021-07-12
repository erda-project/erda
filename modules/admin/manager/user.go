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
	"net/http"

	"github.com/gorilla/schema"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
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
	resp, err := am.bundle.ListUsers(listReq)
	if err != nil {
		return apierrors.ErrListUser.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(resp)
}

func (am *AdminManager) SearchUser(ctx context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {
	queryDecoder := schema.NewDecoder()
	queryDecoder.IgnoreUnknownKeys(true)

	resp, err := am.bundle.SearchUser(req.URL.Query())
	if err != nil {
		return apierrors.ErrListUser.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(resp)
}
