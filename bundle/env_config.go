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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httputil"
)

// CreateNamespace 创建配置中心 Namespace
func (b *Bundle) CreateNamespace(createReq apistructs.NamespaceCreateRequest) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var createResp apistructs.NamespaceCreateResponse
	resp, err := hc.Post(host).Path("/api/config/namespace").
		Header(httputil.InternalHeader, "bundle").
		JSONBody(createReq).
		Do().JSON(&createResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !createResp.Success {
		return toAPIError(resp.StatusCode(), createResp.Error)
	}
	return nil
}

// CreateNamespaceRelations 创建 Namespace 关联关系
func (b *Bundle) CreateNamespaceRelations(createReq apistructs.NamespaceRelationCreateRequest) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var createResp apistructs.NamespaceRelationCreateResponse
	resp, err := hc.Post(host).Path("/api/config/namespace/relation").
		Header(httputil.InternalHeader, "bundle").
		JSONBody(createReq).
		Do().JSON(&createResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !createResp.Success {
		return toAPIError(resp.StatusCode(), createResp.Error)
	}
	return nil
}

// DeleteNamespace 删除配置中心 Namespace
func (b *Bundle) DeleteNamespace(name string) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var deleteResp apistructs.NamespaceDeleteResponse
	resp, err := hc.Delete(host).Path("/api/config/namespace").
		Header(httputil.InternalHeader, "bundle").
		Param("name", name).Do().JSON(&deleteResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !deleteResp.Success {
		return toAPIError(resp.StatusCode(), deleteResp.Error)
	}
	return nil
}

// DeleteNamespaceRelation 删除配置中心 namespace relation
func (b *Bundle) DeleteNamespaceRelation(defaultNamespace string) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var deleteResp apistructs.NamespaceRelationDeleteResponse
	resp, err := hc.Delete(host).Path("/api/config/namespace/relation").
		Header(httputil.InternalHeader, "bundle").
		Param("default_namespace", defaultNamespace).Do().JSON(&deleteResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !deleteResp.Success {
		return toAPIError(resp.StatusCode(), deleteResp.Error)
	}
	return nil
}

// FetchDeploymentConfig 通过 namespace 查询部署配置
// return (ENV, FILE, error)
func (b *Bundle) FetchDeploymentConfig(namespace string) (map[string]string, map[string]string, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, nil, err
	}
	hc := b.hc

	var fetchResp apistructs.EnvConfigFetchResponse
	resp, err := hc.Get(host).Path("/api/config/deployment").
		Param("namespace_name", namespace).Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}
	envs := make(map[string]string)
	files := make(map[string]string)
	for _, c := range fetchResp.Data {
		if c.ConfigType == "FILE" {
			files[c.Key] = c.Value
		} else {
			envs[c.Key] = c.Value
		}

	}
	return envs, files, nil
}

// FetchNamespaceConfig 通过 namespace 查询配置.
func (b *Bundle) FetchNamespaceConfig(fetchReq apistructs.EnvConfigFetchRequest) (map[string]string, error) {
	host, err := b.urls.CMDB()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	// auto create namespace if not exist to avoid error [namespace.not.exist].
	if fetchReq.AutoCreateIfNotExist {
		if fetchReq.CreateReq.ProjectID <= 0 {
			return nil, errors.Errorf("invalid projectID: %d", fetchReq.CreateReq.ProjectID)
		}
		if err := b.CreateNamespace(fetchReq.CreateReq); err != nil {
			return nil, err
		}
	}

	var fetchResp apistructs.EnvConfigFetchResponse
	resp, err := hc.Get(host).Path("/api/config").
		Header(httputil.InternalHeader, "bundle").
		Param("namespace_name", fetchReq.Namespace).Param("decrypt", fmt.Sprintf("%t", fetchReq.Decrypt)).
		Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	kvs := make(map[string]string, 0)
	for _, item := range fetchResp.Data {
		if item.Key != "" {
			kvs[item.Key] = item.Value
		}
	}
	return kvs, nil
}

// AddOrUpdateNamespaceConfig 新增或更新 namespace config .
func (b *Bundle) AddOrUpdateNamespaceConfig(namespace string, items []apistructs.EnvConfig, encrypt bool) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var updateResp apistructs.Header
	resp, err := hc.Post(host).
		Path("/api/config").
		Header(httputil.InternalHeader, "bundle").
		Param("namespace_name", namespace).
		Param("encrypt", fmt.Sprintf("%t", encrypt)).
		JSONBody(&apistructs.EnvConfigAddOrUpdateRequest{Configs: items}).
		Do().JSON(&updateResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !updateResp.Success {
		return toAPIError(resp.StatusCode(), updateResp.Error)
	}

	return nil
}

// DeleteNamespaceConfig 删除 namespace 配置.
func (b *Bundle) DeleteNamespaceConfig(namespace string, key string) error {
	host, err := b.urls.CMDB()
	if err != nil {
		return err
	}
	hc := b.hc

	var updateResp apistructs.Header
	resp, err := hc.Delete(host).
		Path("/api/config").
		Header(httputil.InternalHeader, "bundle").
		Param("namespace_name", namespace).
		Param("key", key).
		Do().JSON(&updateResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !updateResp.Success {
		return toAPIError(resp.StatusCode(), updateResp.Error)
	}

	return nil
}
