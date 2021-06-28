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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (e *Endpoints) CreateAkSks(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.AkSkCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateAkSk.InvalidParameter(err).ToResp(), nil
	}
	obj, err := e.aksk.CreateAkSk(ctx, req)
	if err != nil {
		return apierrors.ErrCreateAkSk.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(obj)
}

func (e *Endpoints) GetAkSkByAk(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	ak, ok := vars["ak"]
	if !ok {
		return apierrors.ErrGetAkSk.InvalidParameter(errors.New("no path param ak")).ToResp(), nil
	}
	obj, err := e.aksk.GetAkSkByAk(ctx, ak)
	if err != nil {
		return apierrors.ErrGetAkSk.NotFound().ToResp(), nil
	}

	return httpserver.OkResp(obj)
}

func (e *Endpoints) DeleteAkSkByAk(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	ak, ok := vars["ak"]
	if !ok {
		return apierrors.ErrDeleteAkSk.InvalidParameter(errors.New("no path param ak")).ToResp(), nil
	}
	err := e.aksk.DeleteAkSkByAk(ctx, ak)
	if err != nil {
		return apierrors.ErrDeleteAkSk.NotFound().ToResp(), nil
	}
	return httpserver.OkResp(nil)
}
