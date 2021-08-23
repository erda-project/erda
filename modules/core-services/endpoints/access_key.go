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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
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

func getAccessKeyListParam(r *http.Request) (apistructs.AccessKeyListQueryRequest, error) {
	q, res := r.URL.Query(), apistructs.AccessKeyListQueryRequest{}

	if v := q.Get("isSystem"); v != "" {
		val, err := strconv.ParseBool(v)
		if err != nil {
			return apistructs.AccessKeyListQueryRequest{}, err
		}
		res.IsSystem = &val
	}

	if v := q.Get("status"); v != "" {
		res.Status = v
	}

	if v := q.Get("subjectType"); v != "" {
		res.SubjectType = v
	}

	if v := q.Get("subject"); v != "" {
		res.Subject = v
	}
	return res, nil
}

func (e *Endpoints) ListAccessKeys(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	req, err := getAccessKeyListParam(r)
	if err != nil {
		return apierrors.ErrGetAccessKey.InvalidParameter(err).ToResp(), nil
	}
	obj, err := e.accesskey.ListAccessKey(ctx, req)
	if err != nil {
		return apierrors.ErrGetAccessKey.InternalError(err).ToResp(), nil
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
