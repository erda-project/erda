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
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
)

func (e *Endpoints) GetFileRecord(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetAutoTestSceneSet.NotLogin().ToResp(), nil
	}

	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return apierrors.ErrGetFileRecord.InvalidParameter("id").ToResp(), nil
	}

	record, err := e.testcase.GetFileRecord(id)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(record, []string{record.OperatorID})
}

// Get File Records
func (e *Endpoints) GetFileRecordsByProjectId(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListFileRecord.NotLogin().ToResp(), nil
	}

	var req apistructs.ListTestFileRecordsRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrListFileRecord.InvalidParameter(err).ToResp(), nil
	}

	rsp, operators, err := e.testcase.ListFileRecordsByProject(req)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(rsp, operators)
}
