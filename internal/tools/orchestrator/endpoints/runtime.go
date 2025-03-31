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

package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"

	"github.com/erda-project/erda-infra/pkg/transport"
	"github.com/erda-project/erda-proto-go/core/dicehub/release/pb"
	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/internal/pkg/user"
	"github.com/erda-project/erda/internal/tools/orchestrator/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Endpoints) GetRuntimeSpec(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrGetRuntime.InvalidParameter(err).ToResp(), nil
	}
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrGetRuntime.NotLogin().ToResp(), nil
	}
	v := vars["runtimeID"]
	runtimeID, err := strutil.Atoi64(v)
	if err != nil {
		return apierrors.ErrGetRuntime.InvalidParameter(strutil.Concat("runtimeID: ", v)).ToResp(), nil
	}
	data, err := e.runtime.GetSpec(userID, orgID, uint64(runtimeID))
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(data)
}

func getOrgID(r *http.Request) (uint64, error) {
	v := r.Header.Get("Org-ID")
	if v == "" {
		return 0, fmt.Errorf("Org-ID")
	}
	orgID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0, err
	}
	return orgID, nil
}

// Deprecated
// ReferCluster 查看 runtime & addon 是否有使用集群
func (e *Endpoints) ReferCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrReferRuntime.NotLogin().ToResp(), nil
	}

	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrReferRuntime.InvalidParameter(err).ToResp(), nil
	}

	clusterName := r.URL.Query().Get("cluster")
	referred := e.runtime.ReferCluster(clusterName, orgID)

	return httpserver.OkResp(referred)
}

// OrgcenterJobLogs 包装数据中心--->任务列表 日志
func (e *Endpoints) OrgcenterJobLogs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrGetRuntime.InvalidParameter(err).ToResp(), nil
	}
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrCreateAddon.NotLogin().ToResp(), nil
	}
	jobID := r.URL.Query().Get("id")
	if jobID == "" {
		return apierrors.ErrGetRuntime.MissingParameter("id").ToResp(), nil
	}
	clusterName := r.URL.Query().Get("clusterName")
	if clusterName == "" {
		return apierrors.ErrGetRuntime.MissingParameter("clusterName").ToResp(), nil
	}
	result, err := e.runtime.OrgJobLogs(ctx, userID, orgID, r.Header.Get("org"), jobID, clusterName, r.URL.Query())
	if err != nil {
		return apierrors.ErrGetRuntime.InvalidParameter(strutil.Concat("jobID: ", jobID)).ToResp(), nil
	}
	return httpserver.OkResp(result)
}

// GetAppWorkspaceReleases 获取应用某一环境下可部署的 releases
func (e *Endpoints) GetAppWorkspaceReleases(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrGetAppWorkspaceReleases.NotLogin().ToResp(), nil
	}

	// parse query params
	var req apistructs.AppWorkspaceReleasesGetRequest
	if err := queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrGetAppWorkspaceReleases.InvalidParameter(err).ToResp(), nil
	}
	if !req.Workspace.Deployable() {
		return apierrors.ErrGetAppWorkspaceReleases.InvalidParameter(errors.Errorf("invalid workspace: %s", req.Workspace)).ToResp(), nil
	}

	// check app permission
	access, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.AppScope,
		ScopeID:  req.AppID,
		Resource: apistructs.AppResource,
		Action:   apistructs.GetAction,
	})
	if err != nil || !access.Access {
		return apierrors.ErrGetAppWorkspaceReleases.AccessDenied().ToResp(), nil
	}

	// get workspace cluster
	app, err := e.bdl.GetApp(req.AppID)
	if err != nil {
		return apierrors.ErrGetAppWorkspaceReleases.InvalidParameter(err).ToResp(), nil
	}
	var clusterName string
	for _, workspace := range app.Workspaces {
		if workspace.Workspace == req.Workspace.String() {
			clusterName = workspace.ClusterName
			break
		}
	}
	if clusterName == "" {
		return apierrors.ErrGetAppWorkspaceReleases.InternalError(errors.Errorf("not found cluster for %s", req.Workspace)).ToResp(), nil
	}

	branches, err := e.bdl.GetAllValidBranchWorkspace(req.AppID, string(userID))
	if err != nil {
		return apierrors.ErrGetAppWorkspaceReleases.InternalError(err).ToResp(), nil
	}

	// 返回当前环境所有 branch 可部署的 release 列表
	result := make(apistructs.AppWorkspaceReleasesGetResponseData)
	for _, branch := range branches {
		matchArtifactWorkspace := false
		for _, artifactWorkspace := range strutil.Split(branch.ArtifactWorkspace, ",", true) {
			if artifactWorkspace == req.Workspace.String() {
				matchArtifactWorkspace = true
				break
			}
		}
		if !matchArtifactWorkspace {
			continue
		}
		ctx = transport.WithHeader(ctx, metadata.New(map[string]string{httputil.InternalHeader: "true"}))
		releases, err := e.releaseSvc.ListRelease(ctx, &pb.ReleaseListRequest{
			Branch:                       branch.Name,
			CrossClusterOrSpecifyCluster: clusterName,
			ApplicationID:                []string{strconv.FormatUint(req.AppID, 10)},
			PageSize:                     5,
			PageNo:                       1,
		})
		if err != nil {
			return apierrors.ErrGetAppWorkspaceReleases.InternalError(err).ToResp(), nil
		}
		result[branch.Name] = releases.Data
	}

	return httpserver.OkResp(result)
}

func (e *Endpoints) KillPod(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.RuntimeKillPodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrKillPod.InvalidParameter("req body").ToResp(), nil
	}

	if err := e.runtime.KillPod(req.RuntimeID, req.PodName); err != nil {
		return apierrors.ErrKillPod.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(nil)
}
