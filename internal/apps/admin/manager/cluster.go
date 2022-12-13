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

package manager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"

	"github.com/erda-project/erda-infra/pkg/transport"
	clusterpb "github.com/erda-project/erda-proto-go/core/clustermanager/cluster/pb"
	orgpb "github.com/erda-project/erda-proto-go/core/org/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/apps/admin/apierrors"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/pkg/common/apis"
	"github.com/erda-project/erda/pkg/discover"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
)

type (
	ClusterInfoResponse struct {
		Id             int32                 `json:"id"`
		Name           string                `json:"name"`
		DisplayName    string                `json:"displayName"`
		Type           string                `json:"type"`
		Description    string                `json:"description"`
		WildcardDomain string                `json:"wildcardDomain"`
		SchedConfig    *SchedConfigResponse  `json:"schedConfig"`
		ManageConfig   *ManageConfigResponse `json:"manageConfig"`
	}
	SchedConfigResponse struct {
		CpuSubscribeRatio        string `json:"cpuSubscribeRatio"`
		DevCPUSubscribeRatio     string `json:"devCPUSubscribeRatio"`
		TestCPUSubscribeRatio    string `json:"testCPUSubscribeRatio"`
		StagingCPUSubscribeRatio string `json:"stagingCPUSubscribeRatio"`
	}
	ManageConfigResponse struct {
		Address          string `json:"address"`
		CredentialSource string `json:"credentialSource"`
	}
)

func (am *AdminManager) AppendClusterEndpoint() {
	am.endpoints = append(am.endpoints, []httpserver.Endpoint{
		{Path: "/api/clusters", Method: http.MethodGet, Handler: am.ListCluster},
		{Path: "/api/clusters/{clusterName}", Method: http.MethodGet, Handler: am.InspectCluster},
		{Path: "/api/clusters/actions/dereference", Method: http.MethodPut, Handler: am.DereferenceCluster},
	}...)
}

func (am *AdminManager) ListCluster(ctx context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {
	// get user info
	userID := req.Header.Get(httputil.UserHeader)
	if USERID(userID).Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(ErrInvalidUserID).ToResp(), nil
	}

	// get org
	orgID, err := GetOrgID(req)
	if err != nil {
		return apierrors.ErrListApprove.InvalidParameter(err).ToResp(), nil
	}
	orgIDStr := orgIDUint64ToStr(orgID)

	// get cluster type
	clusterType := req.URL.Query().Get("clusterType")

	// check permission
	if err = PermissionCheck(am.bundle, userID, orgIDStr, "", apistructs.GetAction); err != nil {
		logrus.Errorf("list cluster failed, error: %v", err)
		return apierrors.ErrListCluster.InternalError(err).ToResp(), nil
	}

	// get cluster list
	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "admin"}))
	// list clusters with clusterType
	resp, err := am.clusterSvc.ListCluster(ctx, &clusterpb.ListClusterRequest{
		OrgID:       uint32(orgID),
		ClusterType: clusterType,
	})
	if err != nil {
		return apierrors.ErrListCluster.InternalError(err).ToResp(), nil
	}

	// use sys symbol if you use admin user and return all cluster info
	// else return cluster with org relation
	if req.URL.Query().Get("sys") == "" {
		orgResp, err := am.org.GetOrgClusterRelationsByOrg(
			apis.WithInternalClientContext(ctx, discover.SvcAdmin),
			&orgpb.GetOrgClusterRelationsByOrgRequest{OrgID: orgIDStr},
		)
		if err != nil {
			return apierrors.ErrListCluster.InternalError(err).ToResp(), nil
		}

		for _, cluster := range resp.Data {
			for _, relate := range orgResp.Data {
				if relate.ClusterID == uint64(cluster.Id) && relate.OrgID == orgID {
					cluster.IsRelation = "Y"
				}
			}
		}
	}

	return httpserver.OkResp(convertClusterInfoListToResponse(resp.Data))
}

func convertClusterInfoListToResponse(clusters []*clusterpb.ClusterInfo) []*ClusterInfoResponse {
	clusterInfoResp := make([]*ClusterInfoResponse, 0, len(clusters))
	for _, cluster := range clusters {
		clusterInfoResp = append(clusterInfoResp, convertClusterInfoToResponse(cluster))
	}

	return clusterInfoResp
}

func convertClusterInfoToResponse(cluster *clusterpb.ClusterInfo) *ClusterInfoResponse {
	if cluster == nil {
		return &ClusterInfoResponse{}
	}

	resp := &ClusterInfoResponse{
		Id:             cluster.Id,
		Name:           cluster.Name,
		DisplayName:    cluster.DisplayName,
		Type:           cluster.Type,
		Description:    cluster.Description,
		WildcardDomain: cluster.WildcardDomain,
		SchedConfig:    &SchedConfigResponse{},
		ManageConfig:   &ManageConfigResponse{},
	}

	if cluster.SchedConfig != nil {
		resp.SchedConfig = &SchedConfigResponse{
			CpuSubscribeRatio:        cluster.SchedConfig.CpuSubscribeRatio,
			DevCPUSubscribeRatio:     cluster.SchedConfig.DevCPUSubscribeRatio,
			TestCPUSubscribeRatio:    cluster.SchedConfig.TestCPUSubscribeRatio,
			StagingCPUSubscribeRatio: cluster.SchedConfig.StagingCPUSubscribeRatio,
		}
	}

	if cluster.ManageConfig != nil {
		resp.ManageConfig = &ManageConfigResponse{
			Address:          cluster.ManageConfig.Address,
			CredentialSource: cluster.ManageConfig.CredentialSource,
		}
	}

	return resp
}

func (am *AdminManager) DereferenceCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrDereferenceCluster.NotLogin().ToResp(), nil
	}

	orgID, err := GetOrgIDStr(r)
	if err != nil {
		return apierrors.ErrListCluster.InvalidParameter(err).ToResp(), nil
	}

	clusterName := r.URL.Query().Get("clusterName")
	if clusterName == "" {
		return apierrors.ErrDereferenceCluster.MissingParameter("clusterName").ToResp(), nil
	}

	referenceResp, err := am.bundle.FindClusterResource(clusterName, orgID)
	if err != nil {
		return apierrors.ErrDereferenceCluster.InternalError(err).ToResp(), nil
	}
	if referenceResp.AddonReference > 0 || referenceResp.ServiceReference > 0 {
		return apierrors.ErrDereferenceCluster.InternalError(errors.Errorf("集群中存在未清理的Addon或Service，请清理后再执行")).ToResp(), nil
	}

	clusterResp, err := am.org.DereferenceCluster(apis.WithUserIDContext(apis.WithInternalClientContext(ctx, discover.SvcAdmin), userID.String()), &orgpb.DereferenceClusterRequest{
		OrgID:       orgID,
		ClusterName: clusterName,
	})
	if err != nil {
		return apierrors.ErrDereferenceCluster.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(clusterResp.Data)
}

func (am *AdminManager) InspectCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	clusterName := vars["clusterName"]
	if clusterName == "" {
		return apierrors.ErrGetCluster.MissingParameter("clusterName").ToResp(), nil
	}

	// get user info
	userID := r.Header.Get(httputil.UserHeader)
	if USERID(userID).Invalid() {
		return apierrors.ErrGetCluster.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	// get org
	orgIDStr, err := GetOrgIDStr(r)
	if err != nil {
		return apierrors.ErrGetCluster.InvalidParameter(err).ToResp(), nil
	}

	// check permission
	if err := PermissionCheck(am.bundle, userID, orgIDStr, "", apistructs.GetAction); err != nil {
		logrus.Errorf("list cluster failed, error: %v", err)
		return apierrors.ErrGetCluster.InternalError(err).ToResp(), nil
	}

	ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "admin"}))
	resp, err := am.clusterSvc.GetCluster(ctx, &clusterpb.GetClusterRequest{IdOrName: clusterName})
	if err != nil {
		return apierrors.ErrGetCluster.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(convertClusterInfoToResponse(resp.Data))
}
