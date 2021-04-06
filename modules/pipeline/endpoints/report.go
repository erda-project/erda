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
	"fmt"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

func (e *Endpoints) queryPipelineReportSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	// pipelineID
	pipelineID, err := strconv.ParseUint(vars[pathPipelineID], 10, 64)
	if err != nil {
		return apierrors.ErrQueryPipelineReportSet.InvalidParameter(fmt.Errorf("invalid pipelineID: %s", vars[pathPipelineID])).ToResp(), nil
	}

	// identity
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrQueryPipelineReportSet.AccessDenied().ToResp(), nil
	}
	_ = identityInfo

	// query
	set, err := e.reportSvc.GetPipelineReportSet(pipelineID, r.URL.Query()["type"]...)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(set)
}

func (e *Endpoints) pagingPipelineReportSets(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	// 请求参数
	var req apistructs.PipelineReportSetPagingRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrPagingPipelineReports.InvalidParameter(err).ToResp(), nil
	}

	// identity
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrPagingPipelineReports.AccessDenied().ToResp(), nil
	}
	_ = identityInfo

	// query
	pagingResult, err := e.reportSvc.PagingPipelineReportSets(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(pagingResult)
}
