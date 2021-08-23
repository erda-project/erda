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
	"net/url"
	"strings"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// api docs -> https://yuque.antfin-inc.com/terminus_paas_dev/middleware/iggk1u

// UnregisterAddon 注销 runtime addon
func (b *Bundle) AddonMetrics(path string, paramValues url.Values) (map[string]interface{}, error) {
	host, err := b.urls.Monitor()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var data map[string]interface{}
	r, err := hc.Get(host).
		Path(strings.Replace(path, "/api/metrics/charts", "/api/metrics", 1)).
		Params(paramValues).
		Do().
		JSON(&data)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), apistructs.ErrorResponse{Msg: "request addon metrics fail"})
	}
	return data, nil
}

// UnregisterAddon 注销 runtime addon
func (b *Bundle) ProjectResource(projectIDs []uint64) (*apistructs.ProjectResourceResponse, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var data apistructs.ProjectResourceResponse
	r, err := hc.Post(host).
		Path("/api/projects/resource").
		JSONBody(projectIDs).
		Do().
		JSON(&data)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), apistructs.ErrorResponse{Msg: "request project resource fail"})
	}
	return &data, nil
}

// ListByAddonName 通过addonName查询数据
func (b *Bundle) ListByAddonName(addonName, projectID, workspace string) (*apistructs.AddonNameResponse, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var data apistructs.AddonNameResponse
	r, err := hc.Get(host).
		Path(fmt.Sprintf("/api/addons/types/%s", addonName)).
		Param("projectId", projectID).
		Param("workspace", workspace).
		Do().
		JSON(&data)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), apistructs.ErrorResponse{Msg: "request project resource fail"})
	}
	return &data, nil
}

// ListAddonByRuntimeID 通过RuntimeID查询addon
func (b *Bundle) ListAddonByRuntimeID(runtimeID string) (*apistructs.AddonListResponse, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var data apistructs.AddonListResponse
	r, err := hc.Get(host).
		Path(fmt.Sprintf("/api/addons")).Header(httputil.InternalHeader, "bundle").
		Header(httputil.UserHeader, "bundle").Header(httputil.OrgHeader, "bundle").
		Param("type", "runtime").
		Param("value", runtimeID).
		Do().
		JSON(&data)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), data.Error)
	}
	return &data, nil
}

func (b *Bundle) AddonConfigCallback(addonID string, req apistructs.AddonConfigCallBackResponse) (*apistructs.PostAddonConfigCallBackResponse, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	path := fmt.Sprintf("/api/addon-platform/addons/%s/config", addonID)
	var data apistructs.PostAddonConfigCallBackResponse
	r, err := hc.Post(host).
		Path(path).
		JSONBody(req).
		Do().
		JSON(&data)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), apistructs.ErrorResponse{Msg: "add config callback failed"})
	}
	return &data, nil
}

func (b *Bundle) AddonConfigCallbackProvison(addonID string, req apistructs.AddonCreateCallBackResponse) (*apistructs.PostAddonConfigCallBackResponse, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	path := fmt.Sprintf("/api/addon-platform/addons/%s/action/provision", addonID)
	var data apistructs.PostAddonConfigCallBackResponse
	r, err := hc.Post(host).
		Path(path).
		JSONBody(req).
		Do().
		JSON(&data)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), apistructs.ErrorResponse{Msg: "add config callback failed"})
	}
	return &data, nil
}

func (b *Bundle) FindClusterResource(clusterName, orgID string) (*apistructs.ResourceReferenceData, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var resp apistructs.ResourceReferenceResp
	r, err := hc.Get(host).
		Path("/api/resources/reference").
		Param("clusterName", clusterName).
		Param("orgId", orgID).
		Do().
		JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), apistructs.ErrorResponse{Msg: "find cluster resource failed"})
	}
	return &(resp.Data), nil
}

func (b *Bundle) GetAddon(addonid string) (*apistructs.AddonFetchResponseData, error) {
	host, err := b.urls.Orchestrator()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var resp apistructs.AddonFetchResponse
	r, err := hc.Get(host).
		Path(fmt.Sprintf("/api/addon/%s", addonid)).
		Do().
		JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() {
		return nil, toAPIError(r.StatusCode(), apistructs.ErrorResponse{Msg: "get addon failed"})
	}
	return &resp.Data, nil
}
