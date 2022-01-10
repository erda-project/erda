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

package bundle

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (b *Bundle) GetRuntimes(name, applicationId, workspace, orgID, userID string) ([]apistructs.RuntimeSummaryDTO, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp struct {
		apistructs.Header
		Data []apistructs.RuntimeSummaryDTO
	}

	resp, err := hc.Get(host).
		Path(fmt.Sprintf("/api/runtimes?name=%s&applicationId=%s&workspace=%s", name, applicationId, workspace)).
		Header(httputil.OrgHeader, orgID).
		Header(httputil.UserHeader, userID).
		Do().JSON(&rsp)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}
	if len(rsp.Data) == 0 {
		return nil, nil
	}
	return rsp.Data, nil
}

func (b *Bundle) CreateRuntime(req apistructs.RuntimeCreateRequest, orgID uint64, userID string) (*apistructs.DeploymentCreateResponseDTO, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp struct {
		apistructs.Header
		Data apistructs.DeploymentCreateResponseDTO
	}

	resp, err := hc.Post(host).Path(fmt.Sprintf("/api/runtimes")).
		Header(httputil.OrgHeader, strconv.FormatUint(orgID, 10)).
		Header(httputil.UserHeader, userID).
		JSONBody(req).Do().JSON(&rsp)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}

	return &rsp.Data, nil
}

// BatchUpdateScale 表示批量处理 runtimes 的 scale 操作
// action 表示 scale 的操作，可以是如下 3 个取值
// 空值
// scaleUp    表示恢复之前已经停止的 runtimes
// scaleDown  表示停止之前已经在运行的 runtimes
func (b *Bundle) BatchUpdateScale(req apistructs.RuntimeScaleRecords, orgID uint64, userID, action string) (apistructs.BatchRuntimeScaleResults, error) {
	if action != apistructs.ScaleActionDown && action != apistructs.ScaleActionUp && action != "" {
		return apistructs.BatchRuntimeScaleResults{}, errors.Errorf("value of %s is invalid scale action, valid value for BatchUpdate_Scale is '%s' or '%s' or ''", apistructs.ScaleAction, apistructs.ScaleActionUp, apistructs.ScaleActionDown)
	}

	data, err := b.batchProcessRuntimes(req, orgID, userID, action)
	if err != nil {
		return apistructs.BatchRuntimeScaleResults{}, apierrors.ErrInvoke.InternalError(err)
	}

	batchScales, ok := data.(apistructs.BatchRuntimeScaleResults)
	if !ok {
		return apistructs.BatchRuntimeScaleResults{}, apierrors.ErrInvoke.InternalError(err)
	}

	return batchScales, nil
}

// BatchUpdateReDeploy 表示批量处理 runtimes 的 redeploy 操作
// action 表示 scale 的操作，仅支持取值 reDeploy
func (b *Bundle) BatchUpdateReDeploy(req apistructs.RuntimeScaleRecords, orgID uint64, userID, action string) (apistructs.BatchRuntimeReDeployResults, error) {
	if action != apistructs.ScaleActionReDeploy {
		return apistructs.BatchRuntimeReDeployResults{}, errors.Errorf("value of %s is invalid scale action, valid value for BatchUpdate_ReDeploy is '%s' ", apistructs.ScaleAction, apistructs.ScaleActionReDeploy)
	}
	data, err := b.batchProcessRuntimes(req, orgID, userID, action)
	if err != nil {
		return apistructs.BatchRuntimeReDeployResults{}, apierrors.ErrInvoke.InternalError(err)
	}

	batchRedeploys, ok := data.(apistructs.BatchRuntimeReDeployResults)
	if !ok {
		return apistructs.BatchRuntimeReDeployResults{}, apierrors.ErrInvoke.InternalError(err)
	}

	return batchRedeploys, nil
}

// BatchUpdateDelete 表示批量处理 runtimes 的 delete 操作
// action 表示 scale 的操作，仅支持取值 delete
func (b *Bundle) BatchUpdateDelete(req apistructs.RuntimeScaleRecords, orgID uint64, userID, action string) (apistructs.BatchRuntimeDeleteResults, error) {
	if action != apistructs.ScaleActionDelete {
		return apistructs.BatchRuntimeDeleteResults{}, errors.Errorf("value of %s is invalid scale action, valid value for BatchUpdate_ReDeploy is '%s' ", apistructs.ScaleAction, apistructs.ScaleActionReDeploy)
	}
	data, err := b.batchProcessRuntimes(req, orgID, userID, action)
	if err != nil {
		return apistructs.BatchRuntimeDeleteResults{}, apierrors.ErrInvoke.InternalError(err)
	}

	batchRedeploys, ok := data.(apistructs.BatchRuntimeDeleteResults)
	if !ok {
		return apistructs.BatchRuntimeDeleteResults{}, apierrors.ErrInvoke.InternalError(err)
	}

	return batchRedeploys, nil
}

// batchProcessRuntimes 批量处理 runtimes 的 scale、redeploy、delete 操作
// action 表示 scale 的操作，可以是如下 5 个取值
// 空值
// scaleUp    表示恢复之前已经停止的 runtimes
// scaleDown  表示停止之前已经在运行的 runtimes
// delete     表示删除 runtimes
// reDeploy   表示重新部署 runtimes
func (b *Bundle) batchProcessRuntimes(req apistructs.RuntimeScaleRecords, orgID uint64, userID, action string) (interface{}, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp struct {
		apistructs.Header
		Data interface{}
	}

	resp, err := hc.Put(host).Path(fmt.Sprintf("/api/runtimes/actions/batch-update-pre-overlay?scale_action=%s", action)).
		Header(httputil.OrgHeader, strconv.FormatUint(orgID, 10)).
		Header(httputil.UserHeader, userID).
		JSONBody(req).Do().JSON(&rsp)

	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}

	return &rsp.Data, nil
}
