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

	"github.com/erda-project/erda/pkg/strutil"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (e *Endpoints) CreateAccessKey(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.AccessKeyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateAccessKey.InvalidParameter(err).ToResp(), nil
	}
	obj, err := e.accesskey.CreateAccessKey(ctx, req)
	if err != nil {
		return apierrors.ErrCreateAccessKey.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(obj)
}

func (e *Endpoints) UpdateAccessKey(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	ak, ok := vars["accessKeyId"]
	if !ok {
		return apierrors.ErrGetAccessKey.InvalidParameter(errors.New("no path param accessKeyId")).ToResp(), nil
	}
	var req apistructs.AccessKeyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateAccessKey.InvalidParameter(err).ToResp(), nil
	}
	if !strutil.InSlice(req.Status, []string{"", apistructs.AccessKeyStatusDisabled, apistructs.AccessKeyStatusActive}) {
		return apierrors.ErrUpdateAccessKey.InvalidParameter(errors.New("status verified failed")).ToResp(), nil
	}

	obj, err := e.accesskey.UpdateAccessKey(ctx, ak, req)
	if err != nil {
		return apierrors.ErrUpdateAccessKey.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(obj)
}

func (e *Endpoints) GetByAccessKeyID(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	ak, ok := vars["accessKeyId"]
	if !ok {
		return apierrors.ErrGetAccessKey.InvalidParameter(errors.New("no path param accessKeyId")).ToResp(), nil
	}
	obj, err := e.accesskey.GetByAccessKeyID(ctx, ak)
	if err != nil {
		return apierrors.ErrGetAccessKey.NotFound().ToResp(), nil
	}

	return httpserver.OkResp(obj)
}

func (e *Endpoints) DeleteByAccessKeyID(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	ak, ok := vars["accessKeyId"]
	if !ok {
		return apierrors.ErrDeleteAccessKey.InvalidParameter(errors.New("no path param accessKeyId")).ToResp(), nil
	}
	err := e.accesskey.DeleteByAccessKeyID(ctx, ak)
	if err != nil {
		return apierrors.ErrDeleteAccessKey.NotFound().ToResp(), nil
	}
	return httpserver.OkResp(nil)
}
