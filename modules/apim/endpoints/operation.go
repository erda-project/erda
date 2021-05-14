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

// method + path 确定一个 operation, operation 即接口

package endpoints

import (
	"context"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/apim/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
)

// 从文档中搜索符合条件的接口列表
func (e *Endpoints) SearchOperations(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.SearchOperations.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.SearchOperations.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	var params apistructs.SearchOperationQueryParameters
	if err = e.queryStringDecoder.Decode(&params, r.URL.Query()); err != nil {
		return apierrors.SearchOperations.InvalidParameter("invalid query parameters").ToResp(), nil
	}
	if params.Keyword == "" {
		return apierrors.SearchOperations.MissingParameter("keyword").ToResp(), nil
	}

	var req = apistructs.SearchOperationsReq{
		OrgID:       orgID,
		Identity:    &identity,
		QueryParams: params,
	}
	data, apiError := e.assetSvc.SearchOperations(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(data)
}

// 查询文档中的接口详情
func (e *Endpoints) GetOperation(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.GetOperation.NotLogin().ToResp(), nil
	}
	orgID, err := user.GetOrgID(r)
	if err != nil {
		return apierrors.GetOperation.MissingParameter(apierrors.MissingOrgID).ToResp(), nil
	}

	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.GetOperation.InvalidParameter("invalid id").NotFound().ToResp(), nil
	}
	var params = apistructs.GetOperationURIParameters{
		ID: id,
	}

	var req = apistructs.GetOperationReq{
		OrgID:     orgID,
		Identity:  &identity,
		URIParams: params,
	}

	data, apiError := e.assetSvc.GetOperation(&req)
	if apiError != nil {
		return apiError.ToResp(), nil
	}

	return httpserver.OkResp(data)
}
