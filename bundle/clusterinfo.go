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
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/strutil"
)

// QueryClusterInfo 查询集群信息
// Deprecated: please do not use this bundle method to get cluster info, use *Bundle.GetCluster instead.
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
