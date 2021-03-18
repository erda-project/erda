package bundle

import (
	"fmt"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
)

// GetCluster 查询集群
func (b *Bundle) GetCluster(idOrName string) (*apistructs.ClusterInfo, error) {
	host, err := b.urls.CMDB()
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
	host, err := b.urls.CMDB()
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

	host, err := b.urls.Ops()
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

// UpdateCluster 更新集群信息
func (b *Bundle) UpdateCluster(req apistructs.ClusterUpdateRequest, header ...http.Header) error {
	host, err := b.urls.CMDB()
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
