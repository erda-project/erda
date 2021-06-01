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

package endpoints

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateCluster 创建集群
func (e *Endpoints) CreateCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {

	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		_, err := user.GetUserID(r)
		if err != nil {
			return apierrors.ErrCreateCluster.NotLogin().ToResp(), nil
		}
	}

	if r.Body == nil {
		return apierrors.ErrCreateCluster.MissingParameter("body").ToResp(), nil
	}
	var req apistructs.ClusterCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateCluster.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", req)

	clusterID, err := e.cluster.CreateWithEvent(&req)
	if err != nil {
		return apierrors.ErrCreateCluster.InvalidParameter(err).ToResp(), nil
	}

	return httpserver.OkResp(clusterID)
}

// UpdateCluster 更新集群
func (e *Endpoints) UpdateCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return apierrors.ErrUpdateCluster.NotLogin().ToResp(), nil
	}

	orgID, err := strutil.Atoi64(orgIDStr)
	if err != nil {
		return apierrors.ErrUpdateCluster.InvalidParameter(err).ToResp(), nil
	}

	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		userID, err := user.GetUserID(r)
		if err != nil {
			return apierrors.ErrUpdateCluster.NotLogin().ToResp(), nil
		}

		// 操作鉴权
		permissionReq := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.OrgScope,
			ScopeID:  uint64(orgID),
			Resource: apistructs.ClusterResource,
			Action:   apistructs.UpdateAction,
		}
		if access, err := e.permission.CheckPermission(&permissionReq); err != nil || !access {
			return apierrors.ErrUpdateCluster.AccessDenied().ToResp(), nil
		}
	}

	if r.Body == nil {
		return apierrors.ErrUpdateCluster.MissingParameter("body").ToResp(), nil
	}
	var req apistructs.ClusterUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateCluster.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", req)

	// TODO 请求body参数合法性检查
	if req.Name == "" {
		return apierrors.ErrUpdateCluster.MissingParameter("name").ToResp(), nil
	}

	clusterID, err := e.cluster.UpdateWithEvent(orgID, &req)
	if err != nil {
		return apierrors.ErrUpdateCluster.InvalidParameter(err).ToResp(), nil
	}

	return httpserver.OkResp(clusterID)
}

// GetCluster 获取集群详情
func (e *Endpoints) GetCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var clientID string
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		_, err := user.GetUserID(r)
		if err != nil {
			clientID = r.Header.Get("Client-ID")
			if !strutil.Contains(clientID, "soldier") {
				return apierrors.ErrGetCluster.NotLogin().ToResp(), nil
			}
		}
	}

	if vars["idOrName"] == "" {
		return apierrors.ErrGetCluster.MissingParameter("idOrName").ToResp(), nil
	}

	cluster, err := e.cluster.GetClusterByIDOrName(vars["idOrName"])
	if err != nil {
		if strutil.Contains(err.Error(), "not found") {
			return apierrors.ErrGetCluster.NotFound().ToResp(), nil
		}
		return apierrors.ErrGetCluster.InternalError(err).ToResp(), nil
	}
	if internalClient == "" {
		orgIDStr := r.Header.Get("Org-ID")
		if orgIDStr == "" {
			return apierrors.ErrGetCluster.NotLogin().ToResp(), nil
		}
		orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
		if err != nil {
			return apierrors.ErrGetCluster.InvalidParameter(err).ToResp(), nil
		}
		clusters, err := e.cluster.ListClusterByOrg(orgID)
		if err != nil {
			return apierrors.ErrGetCluster.InternalError(err).ToResp(), nil
		}

		var exist bool
		for _, v := range *clusters {
			if v.ID == cluster.ID {
				exist = true
				break
			}
		}
		if !exist {
			return apierrors.ErrGetCluster.NotFound().ToResp(), nil
		}
	}

	if clientID == "" { // 只有solider请求才返回密码等敏感信息
		if cluster.System != nil {
			cluster.System.SSH.Password = ""
			cluster.System.SSH.PrivateKey = ""
			cluster.System.SSH.PublicKey = "" //
			cluster.System.Platform.MySQL.Password = ""
		}
		if cluster.SchedConfig != nil {
			cluster.SchedConfig.AuthPassword = ""
			cluster.SchedConfig.ClientCrt = ""
			cluster.SchedConfig.CACrt = ""
			cluster.SchedConfig.ClientKey = ""
			if cluster.SchedConfig.CPUSubscribeRatio == "" {
				cluster.SchedConfig.CPUSubscribeRatio = "1"
			}
		}
		delete(cluster.Settings, "nexusPassword")
		delete(cluster.Config, "nexusPassword")
	}

	return httpserver.OkResp(*cluster)
}

// ListCluster 获取集群列表(可按orgID过滤)
func (e *Endpoints) ListCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		clusters    *[]apistructs.ClusterInfo
		orgID       int64
		userID      string
		clusterType string
		err         error
	)

	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		uid, err := user.GetUserID(r)
		if err != nil {
			return apierrors.ErrListCluster.NotLogin().ToResp(), nil
		}
		userID = uid.String()
	}
	clusterType = r.URL.Query().Get("clusterType")
	orgIDStr := r.URL.Query().Get("orgID")
	if orgIDStr != "" {
		if orgID, err = strutil.Atoi64(orgIDStr); err != nil {
			return apierrors.ErrListCluster.InvalidParameter(err).ToResp(), nil
		}
	}

	if orgID > 0 || clusterType != "" {
		if userID != "" && orgID > 0 { // 外部用户须鉴权
			// 操作鉴权
			req := apistructs.PermissionCheckRequest{
				UserID:   userID,
				Scope:    apistructs.OrgScope,
				ScopeID:  uint64(orgID),
				Resource: apistructs.ClusterResource,
				Action:   apistructs.ListAction,
			}
			if access, err := e.permission.CheckPermission(&req); err != nil || !access {
				return apierrors.ErrListCluster.AccessDenied().ToResp(), nil
			}
		}
		clusters, err = e.cluster.ListClusterByOrgAndType(orgID, clusterType)
	} else {
		if internalClient == "" { // 获取所有集群列表只能走内部调用
			// 操作鉴权
			req := apistructs.PermissionCheckRequest{
				UserID: userID,
			}
			if access, err := e.permission.CheckPermission(&req); err != nil || !access {
				return apierrors.ErrListCluster.AccessDenied().ToResp(), nil
			}
		}
		clusters, err = e.cluster.ListCluster()
	}
	if err != nil {
		return apierrors.ErrListCluster.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(clusters)
}

// DeleteCluster 删除集群
func (e *Endpoints) DeleteCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteCluster.NotLogin().ToResp(), nil
	}
	// 仅内部调用
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrDeleteCluster.AccessDenied().ToResp(), nil
	}

	if err := e.cluster.DeleteWithEvent(vars["clusterName"]); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp("delete succ")
}

// DereferenceCluster 解除关联集群关系
func (e *Endpoints) DereferenceCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrDereferenceCluster.NotLogin().ToResp(), nil
	}
	orgIDStr := r.URL.Query().Get("orgID")
	if orgIDStr == "" {
		return apierrors.ErrDereferenceCluster.MissingParameter("orgID").ToResp(), nil
	}
	var orgID int64
	if orgID, err = strutil.Atoi64(orgIDStr); err != nil {
		return apierrors.ErrListCluster.InvalidParameter(err).ToResp(), nil
	}
	clusterName := r.URL.Query().Get("clusterName")
	if clusterName == "" {
		return apierrors.ErrDereferenceCluster.MissingParameter("clusterName").ToResp(), nil
	}
	req := apistructs.DereferenceClusterRequest{
		OrgID:   orgID,
		Cluster: clusterName,
	}
	if err := e.member.CheckPermission(userID.String(), apistructs.OrgScope, req.OrgID); err != nil {
		return apierrors.ErrDereferenceCluster.InternalError(err).ToResp(), nil
	}
	if err := e.cluster.DereferenceCluster(userID.String(), &req); err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp("delete succ")
}
