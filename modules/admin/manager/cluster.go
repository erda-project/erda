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

package manager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/erda-project/erda/modules/pkg/user"

	"github.com/erda-project/erda/apistructs"

	"github.com/erda-project/erda/modules/admin/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (am *AdminManager) AppendClusterEndpoint() {
	am.endpoints = append(am.endpoints, []httpserver.Endpoint{
		{Path: "/api/clusters", Method: http.MethodGet, Handler: am.ListCluster},
		{Path: "/api/clusters/actions/dereference", Method: http.MethodPut, Handler: am.DereferenceCluster},
	}...)
}

func (am *AdminManager) ListCluster(ctx context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {

	var (
		orgID uint64
		err   error
	)

	orgID, err = GetOrgID(req)
	if err != nil {
		return apierrors.ErrListCluster.InvalidParameter(err).ToResp(), nil
	}

	userID := req.Header.Get("USER-ID")
	id := USERID(userID)
	if id.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	clusterType := req.URL.Query().Get("clusterType")

	clusterRelation, err := am.bundle.GetOrgClusterRelationsByOrg(orgID)
	if err != nil {
		return apierrors.ErrListCluster.InternalError(err).ToResp(), nil
	}

	clusters, err := am.bundle.ListClusters(clusterType, orgID)
	if err != nil {
		return apierrors.ErrListCluster.InternalError(err).ToResp(), nil
	}

	newClusters := []apistructs.ClusterInfo{}
	for _, cluster := range clusters {
		for _, relate := range clusterRelation {
			if relate.ClusterID == uint64(cluster.ID) {
				newClusters = append(newClusters, cluster)
			}
		}
	}

	return httpserver.OkResp(newClusters)
}

func (am *AdminManager) DereferenceCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrDereferenceCluster.NotLogin().ToResp(), nil
	}

	orgID, err := GetOrgID(r)
	if err != nil {
		return apierrors.ErrListCluster.InvalidParameter(err).ToResp(), nil
	}

	clusterName := r.URL.Query().Get("clusterName")
	if clusterName == "" {
		return apierrors.ErrDereferenceCluster.MissingParameter("clusterName").ToResp(), nil
	}

	fmt.Printf("%s %d\n", userID, orgID)

	return httpserver.OkResp("delete successfully")
}
