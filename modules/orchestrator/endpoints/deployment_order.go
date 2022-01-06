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

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/orchestrator/utils"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateDeploymentOrder create deployment order
func (e *Endpoints) CreateDeploymentOrder(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.DeploymentOrderCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// param problem
		logrus.Errorf("failed to parse request body: %v", err)
		return apierrors.ErrCreateRuntime.InvalidParameter("req body").ToResp(), nil
	}

	// TODO: auth

	data, err := e.deploymentOrder.Create(&req)
	if err != nil {
		logrus.Errorf("failed to create deployment order: %v", err)
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(data)
}

// GetDeploymentOrder get deployment order
func (e *Endpoints) GetDeploymentOrder(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orderId := vars["deploymentOrderID"]

	orderDetail, err := e.deploymentOrder.Get(orderId)
	if err != nil {
		logrus.Errorf("failed to get deployment order detail, err: %v", err)
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(orderDetail)
}

// ListDeploymentOrder list deployment order with project id.
func (e *Endpoints) ListDeploymentOrder(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// get page
	pageInfo, err := utils.GetPageInfo(r)
	if err != nil {
		return apierrors.ErrListDeploymentOrder.InvalidParameter(err).ToResp(), nil
	}

	v := r.URL.Query().Get("projectID")
	projectId, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return apierrors.ErrListDeploymentOrder.InvalidParameter(strutil.Concat("projectId: ", v)).ToResp(), nil
	}

	// TODO: auth

	// list deployment orders
	data, err := e.deploymentOrder.List(projectId, &pageInfo)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(data)
}

func (e *Endpoints) DeployDeploymentOrder(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req *apistructs.DeploymentOrderDeployRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// param problem
		logrus.Errorf("failed to parse request body: %v", err)
		return apierrors.ErrCreateRuntime.InvalidParameter("req body").ToResp(), nil
	}

	req.DeploymentOrderId = vars["deploymentOrderID"]

	// TODO: auth

	if err := e.deploymentOrder.Deploy(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

func (e *Endpoints) CancelDeploymentOrder(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req *apistructs.DeploymentOrderCancelRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// param problem
		logrus.Errorf("failed to parse request body: %v", err)
		return apierrors.ErrCreateRuntime.InvalidParameter("req body").ToResp(), nil
	}

	req.DeploymentOrderId = vars["deploymentOrderID"]

	// TODO: auth

	if err := e.deploymentOrder.Cancel(req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(nil)
}

func (e *Endpoints) RenderDeploymentName(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	v := r.URL.Query().Get("releaseID")

	// TODO: auth

	ret, err := e.deploymentOrder.RenderName(v)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(ret)
}
