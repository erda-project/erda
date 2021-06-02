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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateLibReference 创建库引用
func (e *Endpoints) CreateLibReference(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var createReq apistructs.LibReferenceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		return apierrors.ErrCreateLibReference.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateLibReference.NotLogin().ToResp(), nil
	}
	orgID, err := strconv.ParseUint(r.Header.Get(httputil.OrgHeader), 10, 64)
	if err != nil {
		return apierrors.ErrCreateLibReference.InvalidParameter(err).ToResp(), nil
	}
	createReq.IdentityInfo = identityInfo
	createReq.OrgID = orgID

	libReferenceID, err := e.libReference.Create(&createReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(libReferenceID)
}

// DeleteLibReference 删除库引用
func (e *Endpoints) DeleteLibReference(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteLibReference.NotLogin().ToResp(), nil
	}

	libReferenceID, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrDeleteLibReference.MissingParameter("id").ToResp(), nil
	}

	if err := e.libReference.Delete(identityInfo, libReferenceID); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(libReferenceID)
}

// ListLibReference 库引用列表
func (e *Endpoints) ListLibReference(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListLibReference.NotLogin().ToResp(), nil
	}
	var listReq apistructs.LibReferenceListRequest
	if err := e.queryStringDecoder.Decode(&listReq, r.URL.Query()); err != nil {
		return apierrors.ErrListLibReference.InvalidParameter(err).ToResp(), nil
	}
	listReq.IdentityInfo = identityInfo

	listResp, err := e.libReference.List(&listReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	// userIDs
	userIDs := make([]string, 0, len(listResp.List))
	for _, n := range listResp.List {
		userIDs = append(userIDs, n.Creator)
	}
	userIDs = strutil.DedupSlice(userIDs, true)

	return httpserver.OkResp(listResp, userIDs)
}

// ListLibReferenceVersion 移动应用引用库依赖版本列表
func (e *Endpoints) ListLibReferenceVersion(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListLibReferenceVersion.NotLogin().ToResp(), nil
	}

	libIDs := r.URL.Query()["libID"]
	if len(libIDs) == 0 {
		return apierrors.ErrListLibReferenceVersion.MissingParameter("libID").ToResp(), nil
	}

	// fetch publish items
	publishItemReq := &apistructs.QueryPublishItemRequest{
		Ids:      strings.Join(libIDs, ","),
		PageNo:   1,
		PageSize: 10000,
	}
	libVersions, err := e.bdl.QueryPublishItems(publishItemReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	result := make([]apistructs.LibReferenceVersion, 0, len(libVersions.List))
	for _, v := range libVersions.List {
		version := apistructs.LibReferenceVersion{
			LibName: v.Name,
			Version: v.LatestVersion,
		}
		result = append(result, version)
	}

	return httpserver.OkResp(result)
}
