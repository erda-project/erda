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
	"github.com/erda-project/erda/pkg/httputil"
)

// CreateTestAPI 创建API 测试
func (b *Bundle) CreateTestAPI(req apistructs.ApiTestInfo) error {
	host, err := b.urls.QA()
	if err != nil {
		return err
	}
	hc := b.hc

	var createResp apistructs.ApiTestsCreateResponse
	httpResp, err := hc.Post(host).Path("/api/apitests").
		Header(httputil.InternalHeader, "bundle").
		JSONBody(&req).
		Do().JSON(&createResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !createResp.Success {
		return toAPIError(httpResp.StatusCode(), createResp.Error)
	}
	return nil
}

// UpdateTestAPI 更新 API 测试
func (b *Bundle) UpdateTestAPI(req apistructs.ApiTestInfo) error {
	host, err := b.urls.QA()
	if err != nil {
		return err
	}
	hc := b.hc

	var updateResp apistructs.ApiTestsUpdateResponse
	httpResp, err := hc.Put(host).Path(fmt.Sprintf("/api/apitests/%d", req.ApiID)).
		Header(httputil.InternalHeader, "bundle").
		JSONBody(&req).
		Do().JSON(&updateResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !updateResp.Success {
		return toAPIError(httpResp.StatusCode(), updateResp.Error)
	}
	return nil
}

// DeleteTestAPI 删除 API 测试
func (b *Bundle) DeleteTestAPI(apiID int64) error {
	host, err := b.urls.QA()
	if err != nil {
		return err
	}
	hc := b.hc

	var delResp apistructs.ApiTestsDeleteResponse
	httpResp, err := hc.Delete(host).Path(fmt.Sprintf("/api/apitests/%d", apiID)).
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&delResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !delResp.Success {
		return toAPIError(httpResp.StatusCode(), delResp.Error)
	}
	return nil
}

// GetTestAPI 获取指定 API 测试信息
func (b *Bundle) GetTestAPI(apiID int64) (*apistructs.ApiTestInfo, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.ApiTestsGetResponse
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("/api/apitests/%d", apiID)).
		Header(httputil.InternalHeader, "bundle").
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !getResp.Success {
		return nil, toAPIError(httpResp.StatusCode(), getResp.Error)
	}
	return getResp.Data, nil
}

// ListTestAPI 获取 API 测试列表
func (b *Bundle) ListTestAPI(usecaseID int64) ([]*apistructs.ApiTestInfo, error) {
	host, err := b.urls.QA()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.ApiTestsListResponse
	httpResp, err := hc.Get(host).Path("/api/apitests/actions/list-apis").
		Header(httputil.InternalHeader, "bundle").
		Param("usecaseID", strconv.FormatInt(usecaseID, 10)).
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !httpResp.IsOK() || !getResp.Success {
		return nil, toAPIError(httpResp.StatusCode(), getResp.Error)
	}
	return getResp.Data, nil
}
