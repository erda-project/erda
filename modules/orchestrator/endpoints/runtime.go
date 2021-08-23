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
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateRuntime 创建应用实例
func (e *Endpoints) CreateRuntime(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// TODO: 需要等 pipeline action 调用走内网后，再从 header 中取 User-ID (operator)
	var req apistructs.RuntimeCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// param problem
		return apierrors.ErrCreateRuntime.InvalidParameter("req body").ToResp(), nil
	}

	projectInfo, err := e.bdl.GetProject(req.Extra.ProjectID)
	if err != nil {
		return apierrors.ErrCreateRuntime.InternalError(err).ToResp(), nil
	}
	clusterName, ok := projectInfo.ClusterConfig[req.Extra.Workspace]
	if !ok {
		return apierrors.ErrCreateRuntime.InternalError(errors.New("cluster not found")).ToResp(), nil
	}
	req.ClusterName = clusterName

	data, err := e.runtime.Create(user.ID(req.Operator), &req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(data)
}

// CreateRuntimeByRelease 通过releaseId创建runtime
func (e *Endpoints) CreateRuntimeByRelease(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	operator, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrCreateRuntime.NotLogin().ToResp(), nil
	}
	orgid, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrCreateRuntime.InvalidParameter(err).ToResp(), nil
	}
	var req apistructs.RuntimeReleaseCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// param problem
		return apierrors.ErrCreateRuntime.InvalidParameter("req body").ToResp(), nil
	}

	data, err := e.runtime.CreateByReleaseIDPipeline(orgid, operator, &req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(data)
}

func (e *Endpoints) CreateRuntimeByReleaseAction(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	operator, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrCreateRuntime.NotLogin().ToResp(), nil
	}
	var req apistructs.RuntimeReleaseCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// param problem
		return apierrors.ErrCreateRuntime.InvalidParameter("req body").ToResp(), nil
	}
	data, err := e.runtime.CreateByReleaseID(operator, &req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(data)
}

// DeleteRuntime 删除应用实例
func (e *Endpoints) DeleteRuntime(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrDeleteRuntime.InvalidParameter(err).ToResp(), nil
	}
	operator, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrDeleteRuntime.NotLogin().ToResp(), nil
	}
	v := vars["runtimeID"]
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return apierrors.ErrDeleteRuntime.InvalidParameter(strutil.Concat("runtimeID: ", v)).ToResp(), nil
	}
	logrus.Debugf("deleting runtime %d, operator %s", runtimeID, operator)
	runtimedto, err := e.runtime.Delete(operator, orgID, runtimeID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(runtimedto)
}

// StopRuntime 停止应用实例
func (e *Endpoints) StopRuntime(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrDeployRuntime.InvalidParameter(err).ToResp(), nil
	}
	operator, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrDeployRuntime.NotLogin().ToResp(), nil
	}
	v := vars["runtimeID"]
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return apierrors.ErrDeployRuntime.InvalidParameter("runtimeID: " + v).ToResp(), nil
	}
	// TODO: the response should be apistructs.Runtime
	data, err := e.runtime.Redeploy(operator, orgID, runtimeID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(data)
}

// StartRuntime 启动应用实例
func (e *Endpoints) StartRuntime(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	return nil, nil
}

// RestartRuntime 重启应用实例
func (e *Endpoints) RestartRuntime(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	return nil, nil
}

// RedeployRuntime 重新部署, 给 action 调用的 endpoint
func (e *Endpoints) RedeployRuntimeAction(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrDeployRuntime.InvalidParameter(err).ToResp(), nil
	}
	operator, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrDeployRuntime.NotLogin().ToResp(), nil
	}
	v := vars["runtimeID"]
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return apierrors.ErrDeployRuntime.InvalidParameter("runtimeID: " + v).ToResp(), nil
	}
	// TODO: the response should be apistructs.Runtime
	data, err := e.runtime.Redeploy(operator, orgID, runtimeID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(data)
}

// RedeployRuntime 重新部署
func (e *Endpoints) RedeployRuntime(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrDeployRuntime.InvalidParameter(err).ToResp(), nil
	}
	operator, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrDeployRuntime.NotLogin().ToResp(), nil
	}
	v := vars["runtimeID"]
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return apierrors.ErrDeployRuntime.InvalidParameter("runtimeID: " + v).ToResp(), nil
	}
	// TODO: the response should be apistructs.Runtime
	data, err := e.runtime.RedeployPipeline(operator, orgID, runtimeID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(data)
}

func (e *Endpoints) RollbackRuntimeAction(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrRollbackRuntime.InvalidParameter(err).ToResp(), nil
	}
	operator, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrRollbackRuntime.NotLogin().ToResp(), nil
	}
	var req apistructs.RuntimeRollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// param problem
		return apierrors.ErrRollbackRuntime.InvalidParameter(err).ToResp(), nil
	}
	if req.DeploymentID <= 0 {
		return apierrors.ErrRollbackRuntime.InvalidParameter("deploymentID").ToResp(), nil
	}
	v := vars["runtimeID"]
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return apierrors.ErrRollbackRuntime.InvalidParameter(strutil.Concat("runtimeID: ", v)).ToResp(), nil
	}
	data, err := e.runtime.Rollback(operator, orgID, runtimeID, req.DeploymentID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(data)
}

// Rollback 回滚应用实例
func (e *Endpoints) RollbackRuntime(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrRollbackRuntime.InvalidParameter(err).ToResp(), nil
	}
	operator, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrRollbackRuntime.NotLogin().ToResp(), nil
	}
	var req apistructs.RuntimeRollbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// param problem
		return apierrors.ErrRollbackRuntime.InvalidParameter(err).ToResp(), nil
	}
	if req.DeploymentID <= 0 {
		return apierrors.ErrRollbackRuntime.InvalidParameter("deploymentID").ToResp(), nil
	}
	v := vars["runtimeID"]
	runtimeID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return apierrors.ErrRollbackRuntime.InvalidParameter(strutil.Concat("runtimeID: ", v)).ToResp(), nil
	}
	data, err := e.runtime.RollbackPipeline(operator, orgID, runtimeID, req.DeploymentID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(data)
}

// ListRuntimes 查询应用实例列表
func (e *Endpoints) ListRuntimes(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrGetRuntime.InvalidParameter(err).ToResp(), nil
	}
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrGetRuntime.NotLogin().ToResp(), nil
	}
	v := r.URL.Query().Get("applicationId")
	if v == "" {
		v = r.URL.Query().Get("applicationID")
	}
	appID, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return apierrors.ErrListRuntime.InvalidParameter(strutil.Concat("applicationID: ", v)).ToResp(), nil
	}
	workspace := r.URL.Query().Get("workspace")
	name := r.URL.Query().Get("name")
	data, err := e.runtime.List(userID, orgID, appID, workspace, name)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	userIDs := make([]string, 0, len(data))
	for i := range data {
		userIDs = append(userIDs, data[i].LastOperator)
	}
	return httpserver.OkResp(data, userIDs)
}

// GetRuntime 查询应用实例
func (e *Endpoints) GetRuntime(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrGetRuntime.InvalidParameter(err).ToResp(), nil
	}
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrGetRuntime.NotLogin().ToResp(), nil
	}
	var (
		idOrName  = vars["idOrName"]
		appID     = r.URL.Query().Get("applicationId")
		workspace = r.URL.Query().Get("workspace")
	)
	data, err := e.runtime.Get(userID, orgID, idOrName, appID, workspace)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(data)
}

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

// FullGC 触发全量 GC
func (e *Endpoints) FullGC(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	go e.runtime.FullGC()
	return httpserver.OkResp(nil)
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

// ReferCluster 查看 runtime & addon 是否有使用集群
func (e *Endpoints) ReferCluster(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrReferRuntime.NotLogin().ToResp(), nil
	}
	// 仅内部使用
	if !identityInfo.IsInternalClient() {
		return apierrors.ErrReferRuntime.AccessDenied().ToResp(), nil
	}

	clusterName := r.URL.Query().Get("cluster")
	referred := e.runtime.ReferCluster(clusterName)

	return httpserver.OkResp(referred)
}

// runtimeLogs 包装runtime部署日志
func (e *Endpoints) RuntimeLogs(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getOrgID(r)
	if err != nil {
		return apierrors.ErrGetRuntime.InvalidParameter(err).ToResp(), nil
	}
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrCreateAddon.NotLogin().ToResp(), nil
	}
	source := r.URL.Query().Get("source")
	if source == "" {
		return apierrors.ErrGetRuntime.MissingParameter("source").ToResp(), nil
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		return apierrors.ErrGetRuntime.MissingParameter("id").ToResp(), nil
	}
	deploymentID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return apierrors.ErrGetRuntime.InvalidParameter(strutil.Concat("deploymentID: ", id)).ToResp(), nil
	}
	result, err := e.runtime.RuntimeDeployLogs(userID, orgID, deploymentID, r.URL.Query())
	if err != nil {
		return apierrors.ErrGetRuntime.InvalidParameter(strutil.Concat("deploymentID: ", id)).ToResp(), nil
	}
	return httpserver.OkResp(result)

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
	result, err := e.runtime.OrgJobLogs(userID, orgID, jobID, clusterName, r.URL.Query())
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

	branches, err := e.bdl.GetAllValidBranchWorkspace(req.AppID)
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
		releases, err := e.bdl.ListReleases(apistructs.ReleaseListRequest{
			Branch:                       branch.Name,
			CrossClusterOrSpecifyCluster: &clusterName,
			ApplicationID:                int64(req.AppID),
			PageSize:                     5,
			PageNum:                      1,
		})
		if err != nil {
			return apierrors.ErrGetAppWorkspaceReleases.InternalError(err).ToResp(), nil
		}
		result[branch.Name] = releases
	}

	return httpserver.OkResp(result)
}

func (e *Endpoints) KillPod(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.RuntimeKillPodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrKillPod.InvalidParameter("req body").ToResp(), nil
	}
	err := e.runtime.KillPod(req.RuntimeID, req.PodName)
	if err != nil {
		return apierrors.ErrKillPod.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(nil)
}
