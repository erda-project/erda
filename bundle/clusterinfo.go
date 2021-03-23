package bundle

import (
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
)

// QueryClusterInfo 查询集群信息
func (b *Bundle) QueryClusterInfo(name string) (apistructs.ClusterInfoData, error) {
	host, err := b.urls.Scheduler()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var getResp apistructs.ClusterInfoResponse
	resp, err := hc.Get(host).
		Path(strutil.Concat("/api/clusterinfo/", name)).
		Do().
		JSON(&getResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !getResp.Success {
		return nil, toAPIError(resp.StatusCode(), getResp.Error)
	}
	return getResp.Data, nil
}
