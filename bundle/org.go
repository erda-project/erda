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
	"reflect"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httputil"
)

// GetOrg get org by id from core-service.
func (b *Bundle) GetOrg(idOrName interface{}) (*apistructs.OrgDTO, error) {
	host, err := b.urls.ErdaServer()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	if reflect.ValueOf(idOrName).IsZero() {
		return nil, fmt.Errorf("idOrName is empty")
	}
	var fetchResp apistructs.OrgFetchResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/orgs/%v", idOrName)).Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return &fetchResp.Data, nil
}

// GetOrgClusterRelationsByOrg get orgClusters relation by orgID
func (b *Bundle) GetOrgClusterRelationsByOrg(orgID uint64) ([]apistructs.OrgClusterRelationDTO, error) {
	host, err := b.urls.ErdaServer()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var resp apistructs.OrgClusterRelationDTOResponse
	r, err := hc.Get(host).Path(fmt.Sprintf("/api/orgs/clusters/relations/%d", orgID)).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&resp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !r.IsOK() || !resp.Success {
		return nil, toAPIError(r.StatusCode(), resp.Error)
	}

	return resp.Data, nil
}

// CreateOrgClusterRelationsByOrg create orgClusters relation by orgID
func (b *Bundle) CreateOrgClusterRelationsByOrg(clusterName string, userID string, orgID uint64) error {
	host, err := b.urls.ErdaServer()
	if err != nil {
		return err
	}
	hc := b.hc

	org, err := b.GetOrg(orgID)
	if err != nil {
		return err
	}

	var createResp apistructs.OrgClusterRelationDTOCreateResponse

	req := &apistructs.OrgClusterRelationCreateRequest{
		OrgID:       orgID,
		OrgName:     org.Name,
		ClusterName: clusterName,
	}

	resp, err := hc.Post(host).Path("/api/orgs/actions/relate-cluster").
		Header(httputil.UserHeader, userID).
		JSONBody(req).
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
