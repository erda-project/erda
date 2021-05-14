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
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
)

// ListEdasContainers 获取 edas 实例列表 (内部调用)
func (e *Endpoints) ListEdasContainers(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	params, err := getListEdasContainerParams(r)
	if err != nil {
		return apierrors.ErrListInstance.InvalidParameter(err).ToResp(), nil
	}
	containers, err := e.container.ListEdasByParams(params)
	if err != nil {
		return apierrors.ErrListInstance.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(containers)
}

// 获取 edas 容器列表请求参数
func getListEdasContainerParams(r *http.Request) (*apistructs.EdasContainerListRequest, error) {
	var (
		projectID uint64
		appID     uint64
		runtimeID uint64
		err       error
	)
	projectIDStr := r.URL.Query().Get("projectId")
	if projectIDStr != "" {
		projectID, err = strconv.ParseUint(projectIDStr, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	appIDStr := r.URL.Query().Get("appId")
	if appIDStr != "" {
		appID, err = strconv.ParseUint(appIDStr, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	runtimeIDStr := r.URL.Query().Get("runtimeId")
	if runtimeIDStr != "" {
		runtimeID, err = strconv.ParseUint(runtimeIDStr, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	workspace := r.URL.Query().Get("workspace")
	service := r.URL.Query().Get("service")

	edasAppIDs := r.URL.Query()["edasAppId"]

	return &apistructs.EdasContainerListRequest{
		ProjectID:  projectID,
		AppID:      appID,
		RuntimeID:  runtimeID,
		Workspace:  workspace,
		Service:    service,
		EdasAppIDs: edasAppIDs,
	}, nil
}
