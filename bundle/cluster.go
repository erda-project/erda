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
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

// GetCluster 查询集群
func (b *Bundle) GetCluster(idOrName string) (*apistructs.ClusterInfo, error) {
	host, err := b.urls.ClusterManager()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.GetClusterResponse
	resp, err := hc.Get(host).Path(strutil.Concat("/api/clusters/", idOrName)).
		Header("Internal-Client", "bundle").
		Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}
	return &getResp.Data, nil
}

// ListCluster 返回 org 下所有集群; 当 orgID == "" 时，返回所有集群.
func (b *Bundle) ListClusters(clusterType string, orgID ...uint64) ([]apistructs.ClusterInfo, error) {
	host, err := b.urls.ClusterManager()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	req := hc.Get(host).Path("/api/clusters").Header("Internal-Client", "bundle")
	if len(orgID) > 0 {
		req.Param("orgID", fmt.Sprintf("%d", orgID[0]))
	}
	if clusterType != "" {
		req.Param("clusterType", fmt.Sprintf("%s", clusterType))
	}

	var getResp apistructs.ClusterListResponse
	resp, err := req.Do().JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}

	return getResp.Data, nil
}

// DeleteImageManifests 调用 officer 删除 registry image manifests, 真实镜像 blob 由 soldier 删除.
func (b *Bundle) DeleteImageManifests(clusterIDOrName string, images []string) (
	*apistructs.RegistryManifestsRemoveResponseData, error) {

	host, err := b.urls.CMP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var deleteResp apistructs.RegistryManifestsRemoveResponse
	resp, err := hc.Post(host).
		Path(strutil.Concat("/api/clusters/", clusterIDOrName, "/registry/manifests/actions/remove")).
		JSONBody(&apistructs.RegistryManifestsRemoveRequest{Images: images}).
		Do().JSON(&deleteResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !deleteResp.Success {
		return nil, toAPIError(resp.StatusCode(), deleteResp.Error)
	}

	return &deleteResp.Data, nil
}

// UpdateCluster update cluster
func (b *Bundle) UpdateCluster(req apistructs.ClusterUpdateRequest, header ...http.Header) error {
	host, err := b.urls.ClusterManager()
	if err != nil {
		return err
	}
	hc := b.hc

	var updateResp apistructs.ClusterUpdateResponse

	q := hc.Put(host).Path("/api/clusters")
	if len(header) > 0 {
		q.Headers(header[0])
	}
	resp, err := q.JSONBody(req).
		Do().
		JSON(&updateResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !updateResp.Success {
		return toAPIError(resp.StatusCode(), updateResp.Error)
	}
	return nil
}

// CreateCluster Create cluster with event
func (b *Bundle) CreateCluster(req *apistructs.ClusterCreateRequest, header ...http.Header) error {
	host, err := b.urls.ClusterManager()
	if err != nil {
		return err
	}
	hc := b.hc

	var createResp apistructs.ClusterCreateResponse

	q := hc.Post(host).Path("/api/clusters")
	if len(header) > 0 {
		q.Headers(header[0])
	}
	resp, err := q.JSONBody(req).
		Do().
		JSON(&createResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !createResp.Success {
		return toAPIError(resp.StatusCode(), createResp.Error)
	}
	return nil
}

// PatchCluster patch cluster with event
func (b *Bundle) PatchCluster(req *apistructs.ClusterPatchRequest, header ...http.Header) error {
	host, err := b.urls.ClusterManager()
	if err != nil {
		return err
	}
	hc := b.hc

	q := hc.Patch(host).Path("/api/clusters")
	if len(header) > 0 {
		q.Headers(header[0])
	}

	var httpResp httpserver.Resp

	resp, err := q.JSONBody(req).Do().JSON(&httpResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	if !resp.IsOK() || !httpResp.Success {
		return toAPIError(resp.StatusCode(), httpResp.Err)
	}

	return nil
}

// DeleteCluster Delete cluster with event
func (b *Bundle) DeleteCluster(clusterName string, header ...http.Header) error {
	host, err := b.urls.ClusterManager()
	if err != nil {
		return err
	}
	hc := b.hc

	q := hc.Delete(host).Path("/api/clusters/" + clusterName)
	if len(header) > 0 {
		q.Headers(header[0])
	}

	var httpResp httpserver.Resp

	resp, err := q.Do().JSON(&httpResp)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}

	if !resp.IsOK() || !httpResp.Success {
		return toAPIError(resp.StatusCode(), httpResp.Err)
	}

	return nil
}
