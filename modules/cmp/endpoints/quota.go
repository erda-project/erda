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
	"net/http"
	"strconv"

	jsi "github.com/json-iterator/go"
	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/core-services/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
)

var errParamIllegal = errors.New("param illegal")

func (e *Endpoints) GetResourceGauge(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	req := &apistructs.GaugeRequest{}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	userIDStr := r.Header.Get(httputil.UserHeader)
	if orgIDStr == "" {
		return apierrors.ErrFetchOrgResources.NotLogin().ToResp(), nil
	}
	err = jsi.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return httpserver.HTTPResponse{Status: http.StatusInternalServerError}, err
	}
	if req.CpuUnit < 1 || req.MemoryUnit < 1 {
		err = errParamIllegal
		return
	}
	content, err := e.Resource.GetGauge(orgIDStr, userIDStr, req)
	if err != nil {
		return httpserver.HTTPResponse{Status: http.StatusInternalServerError}, err
	}
	return httpserver.HTTPResponse{Status: http.StatusOK, Content: content}, nil
}

func (e *Endpoints) GetResourceClass(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	req := &apistructs.ClassRequest{}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	userIDStr := r.Header.Get(httputil.UserHeader)
	if orgIDStr == "" {
		return apierrors.ErrFetchOrgResources.NotLogin().ToResp(), nil
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrFetchOrgResources.InvalidParameter(err).ToResp(), nil
	}

	err = jsi.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return httpserver.HTTPResponse{Status: http.StatusInternalServerError}, err
	}
	pie, err := e.Resource.GetPie(orgID, userIDStr, req)
	if err != nil {
		return nil, err
	}

	return httpserver.HTTPResponse{Status: http.StatusOK, Content: pie}, nil
}

func (e *Endpoints) GetResourceTable(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	req := &apistructs.TableRequest{}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	userIDStr := r.Header.Get(httputil.UserHeader)
	if orgIDStr == "" {
		return apierrors.ErrFetchOrgResources.NotLogin().ToResp(), nil
	}
	err = jsi.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return httpserver.HTTPResponse{Status: http.StatusInternalServerError}, err
	}
	pie, err := e.Resource.GetTable(orgIDStr, userIDStr, req.ClusterNames, nil, nil)
	if err != nil {
		return nil, err
	}
	return httpserver.HTTPResponse{Status: http.StatusOK, Content: pie}, nil
}

func (e *Endpoints) GetResourceClusterTrend(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	req := &apistructs.TrendRequest{}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	userIDStr := r.Header.Get(httputil.UserHeader)
	if orgIDStr == "" {
		return apierrors.ErrFetchOrgResources.NotLogin().ToResp(), nil
	}
	err = jsi.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return httpserver.HTTPResponse{Status: http.StatusInternalServerError}, err
	}
	pie, err := e.Resource.GetClusterTrend(orgIDStr, userIDStr, req)
	if err != nil {
		return nil, err
	}

	return httpserver.HTTPResponse{Status: http.StatusOK, Content: pie}, nil
}

func (e *Endpoints) GetResourceProjectTrend(ctx context.Context, r *http.Request, vars map[string]string) (resp httpserver.Responser, err error) {
	req := &apistructs.TrendRequest{}
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	userIDStr := r.Header.Get(httputil.UserHeader)
	if orgIDStr == "" {
		return apierrors.ErrFetchOrgResources.NotLogin().ToResp(), nil
	}
	err = jsi.NewDecoder(r.Body).Decode(req)
	if err != nil {
		return httpserver.HTTPResponse{Status: http.StatusInternalServerError}, err
	}
	pie, err := e.Resource.GetProjectTrend(orgIDStr, userIDStr, req)
	if err != nil {
		return nil, err
	}

	return httpserver.HTTPResponse{Status: http.StatusOK, Content: pie}, nil
}
