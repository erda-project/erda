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

package bundle

import (
	"fmt"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// 创建流量入口
func (b *Bundle) CreateEndpoint(orgID, projectID uint64, workspace string, packageDto apistructs.PackageDto) (endpointID string, err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.EndpointInfoResponse
	resp, err := b.hc.Post(host).
		Path("/api/gateway/openapi/packages").
		Header(httputil.InternalHeader, "bundle").
		Param("orgId", strconv.FormatUint(orgID, 10)).
		Param("projectId", strconv.FormatUint(projectID, 10)).
		Param("env", workspace).
		JSONBody(packageDto).Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	endpointID = fetchResp.Data.Id
	return
}

// 更新流量入口
func (b *Bundle) UpdateEndpoint(endpointID string, packageDto apistructs.PackageDto) (err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.Header
	resp, err := b.hc.Patch(host).
		Path(fmt.Sprintf("/api/gateway/openapi/packages/%s", endpointID)).
		Header(httputil.InternalHeader, "bundle").
		JSONBody(packageDto).Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	return
}

// 删除流量入口
func (b *Bundle) DeleteEndpoint(endpointID string) (err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.Header
	resp, err := b.hc.Delete(host).
		Path(fmt.Sprintf("/api/gateway/openapi/packages/%s", endpointID)).
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	return
}

// 获取流量入口详情
func (b *Bundle) GetEndpoint(endpointID string) (dto *apistructs.PackageInfoDto, err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.EndpointInfoResponse
	resp, err := b.hc.Get(host).
		Path(fmt.Sprintf("/api/gateway/openapi/packages/%s", endpointID)).
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	dto = &fetchResp.Data
	return
}

// 创建或更新路由规则
func (b *Bundle) CreateOrUpdateEndpointRootRoute(endpointID, redirectAddr, redirectPath string) (err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.Header
	resp, err := b.hc.Put(host).
		Path(fmt.Sprintf("/api/gateway/openapi/packages/%s/root-api", endpointID)).
		Header(httputil.InternalHeader, "bundle").
		JSONBody(apistructs.OpenapiDto{
			RedirectAddr: redirectAddr,
			RedirectPath: redirectPath,
		}).
		Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	return
}

// 创建调用方
func (b *Bundle) CreateClientConsumer(orgID uint64, clientName string) (dto *apistructs.ClientInfoDto, err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.ClientInfoResponse
	resp, err := b.hc.Post(host).
		Path("/api/gateway/openapi/clients").
		Header(httputil.InternalHeader, "bundle").
		Param("orgId", strconv.FormatUint(orgID, 10)).
		Param("clientName", clientName).
		Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	dto = &fetchResp.Data
	return
}

// 删除调用方
func (b *Bundle) DeleteClientConsumer(clientID string) (err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.Header
	resp, err := b.hc.Delete(host).
		Path(fmt.Sprintf("/api/gateway/openapi/clients/%s", clientID)).
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	return
}

// 获取调用方凭证信息
func (b *Bundle) GetClientCredentials(clientID string) (dto *apistructs.ClientInfoDto, err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.ClientInfoResponse
	resp, err := b.hc.Get(host).
		Path(fmt.Sprintf("/api/gateway/openapi/clients/%s/credentials", clientID)).
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	dto = &fetchResp.Data
	return
}

// 重置调用方密钥
func (b *Bundle) ResetClientCredentials(clientID string) (dto *apistructs.ClientInfoDto, err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.ClientInfoResponse
	resp, err := b.hc.Patch(host).
		Path(fmt.Sprintf("/api/gateway/openapi/clients/%s/credentials", clientID)).
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	dto = &fetchResp.Data
	return
}

// 授权调用方流量入口权限
func (b *Bundle) GrantEndpointToClient(clientID, endpointID string) (err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.Header
	resp, err := b.hc.Post(host).
		Path(fmt.Sprintf("/api/gateway/openapi/clients/%s/packages/%s", clientID, endpointID)).
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	return
}

// 收回调用方流量入口权限
func (b *Bundle) RevokeEndpointFromClient(clientID, endpointID string) (err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.Header
	resp, err := b.hc.Delete(host).
		Path(fmt.Sprintf("/api/gateway/openapi/clients/%s/packages/%s", clientID, endpointID)).
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	return
}

// 获取tenant-group id
func (b *Bundle) GetTenantGroupID(projectID uint64, workspace string) (id string, err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.TenantGroupResponse
	resp, err := b.hc.Get(host).
		Path("/api/gateway/tenant-group").
		Header(httputil.InternalHeader, "bundle").
		Param("projectId", strconv.FormatUint(projectID, 10)).
		Param("env", workspace).
		Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	id = fetchResp.Data
	return
}

// 创建或更新限流规则

func (b *Bundle) CreateOrUpdateClientLimits(clientID, endpointID string, limits []apistructs.LimitType) (err error) {
	host, err := b.urls.Hepa()
	if err != nil {
		return
	}
	var fetchResp apistructs.Header
	resp, err := b.hc.Put(host).
		Path(fmt.Sprintf("/api/gateway/openapi/clients/%s/packages/%s/limits", clientID, endpointID)).
		Header(httputil.InternalHeader, "bundle").
		JSONBody(apistructs.ChangeLimitsReq{Limits: limits}).Do().JSON(&fetchResp)
	if err != nil {
		err = apierrors.ErrInvoke.InternalError(err)
		return
	}
	if !resp.IsOK() || !fetchResp.Success {
		err = toAPIError(resp.StatusCode(), fetchResp.Error)
		return
	}
	return
}
