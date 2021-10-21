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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

func (b *Bundle) FetchQuotaOnClusters(orgID int64, clusterNames []string) (*apistructs.GetQuotaOnClustersResponse, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	type response struct {
		apistructs.Header
		Data *apistructs.GetQuotaOnClustersResponse
	}
	var (
		resp   response
		params = make(url.Values)
	)
	for _, clusterName := range clusterNames {
		params.Add("clusterName", clusterName)
	}
	httpResp, err := hc.Get(host).
		Path(fmt.Sprintf("/api/projects-quota")).
		Params(params).
		Header(httputil.OrgHeader, strconv.FormatInt(orgID, 10)).
		Do().
		JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrListFileRecord.InternalError(err)
	}
	if !httpResp.IsOK() {
		return nil, toAPIError(httpResp.StatusCode(), resp.Error)
	}
	return resp.Data, nil
}

func (b *Bundle) ListOrgNamespace(req *apistructs.OrgClustersNamespaceReq) (*apistructs.OrgClustersNamespaceResp, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc
	var rsp = apistructs.OrgClustersNamespaceResp{}
	httpResp, err := hc.Get(host).Path(fmt.Sprintf("")).Header(httputil.OrgHeader, req.OrgID).
		Do().JSON(req)
	if err != nil {
		return nil, apierrors.ErrListFileRecord.InternalError(err)
	}
	if !httpResp.IsOK() {
		return nil, toAPIError(httpResp.StatusCode(), rsp.Error)
	}
	return &rsp, nil
}
