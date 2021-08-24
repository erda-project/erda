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
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dicehub/errcode"
	"github.com/erda-project/erda/modules/dicehub/response"
	"github.com/erda-project/erda/modules/dicehub/service/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/parser/diceyml"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateRelease POST /api/releases release创建处理
func (e *Endpoints) CreateRelease(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := getPermissionHeader(r)
	if err != nil {
		return apierrors.ErrCreateRelease.NotLogin().ToResp(), nil
	}

	if r.Body == nil {
		return apierrors.ErrCreateRelease.MissingParameter("body").ToResp(), nil
	}
	var releaseRequest apistructs.ReleaseCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&releaseRequest); err != nil {
		return apierrors.ErrCreateRelease.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("creating release...request body: %v\n", releaseRequest)

	if releaseRequest.ReleaseName == "" {
		return apierrors.ErrCreateRelease.MissingParameter("releaseName").ToResp(), nil
	}
	// if releaseRequest.Dice != "" {
	// 	diceStrWithInitContainer, err := e.InjectDiceInitContainer(releaseRequest.Dice)
	// 	if err != nil {
	// 		return apierrors.ErrCreateRelease.InternalError(err).ToResp(), nil
	// 	}
	// 	releaseRequest.Dice = diceStrWithInitContainer
	// }

	// 创建 Release
	releaseID, err := e.release.Create(&releaseRequest)
	if err != nil {
		return apierrors.ErrCreateRelease.InternalError(err).ToResp(), nil
	}

	respBody := &apistructs.ReleaseCreateResponseData{
		ReleaseID: releaseID,
	}

	return httpserver.OkResp(respBody)
}

func (e *Endpoints) InjectDiceInitContainer(diceStr string) (string, error) {
	var diceYml diceyml.Object
	err := yaml.Unmarshal([]byte(diceStr), &diceYml)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal dice yaml")
	}
	ext, err := e.extension.GetExtensionVersion("java-agent", "1.0", true)
	if err != nil {
		return "", errors.Wrap(err, "failed to get java-agent ext")
	}
	var extDice diceyml.Object
	extDiceStr := ext.Dice.(string)
	err = yaml.Unmarshal([]byte(extDiceStr), &extDice)
	if err != nil {
		return "", errors.Wrap(err, "failed to unmarshal ext dice yaml")
	}
	injectJobs := extDice.Jobs
	for _, service := range diceYml.Services {
		if service.Labels != nil && service.Labels["agent"] == "java" {
			shareDirs := []diceyml.SharedDir{}
			if service.Init == nil {
				service.Init = map[string]diceyml.InitContainer{}
			}
			if service.Envs == nil {
				service.Envs = map[string]string{}
			}
			for jobName, job := range injectJobs {
				for _, bind := range job.Binds {
					shareDirs = append(shareDirs, diceyml.SharedDir{SideCar: bind, Main: bind})
				}
				for envName, envValue := range job.Envs {
					service.Envs[envName] = envValue
				}
				service.Init[jobName] = diceyml.InitContainer{
					Image:      job.Image,
					SharedDirs: shareDirs,
					Cmd:        job.Cmd,
					Resources:  job.Resources,
				}
			}
		}
	}
	diceBytes, err := yaml.Marshal(diceYml)
	if err != nil {
		return "", errors.Wrap(err, "failed to marshal diceYaml")
	}
	return string(diceBytes), nil
}

// UpdateRelease PUT /api/releases/<releaseId> release更新处理
func (e *Endpoints) UpdateRelease(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getPermissionHeader(r)
	if err != nil {
		return apierrors.ErrUpdateRelease.NotLogin().ToResp(), nil
	}

	// Check releaseId if exist in path or not
	releaseID := vars["releaseId"]
	if releaseID == "" {
		return apierrors.ErrUpdateRelease.MissingParameter("releaseId").ToResp(), nil
	}

	if r.Body == nil {
		return apierrors.ErrUpdateRelease.MissingParameter("body").ToResp(), nil
	}
	var updateRequest apistructs.ReleaseUpdateRequestData
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		return apierrors.ErrUpdateRelease.InvalidParameter(err).ToResp(), nil
	}
	logrus.Infof("update release info: %+v", updateRequest)

	if err := e.release.Update(orgID, releaseID, &updateRequest); err != nil {
		return apierrors.ErrUpdateRelease.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("Update succ")
}

// UpdateReleaseReference PUT /api/releases/<releaseId> release更新引用处理
func (e *Endpoints) UpdateReleaseReference(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getPermissionHeader(r)
	if err != nil {
		return apierrors.ErrUpdateRelease.NotLogin().ToResp(), nil
	}

	// Check releaseId if exist in path or not
	releaseID := vars["releaseId"]
	if releaseID == "" {
		return apierrors.ErrUpdateRelease.MissingParameter("releaseId").ToResp(), nil
	}

	// Only update reference
	if r.Body == nil {
		return apierrors.ErrUpdateRelease.MissingParameter("body").ToResp(), nil
	}
	var updateReferenceRequest apistructs.ReleaseReferenceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReferenceRequest); err != nil {
		return apierrors.ErrUpdateRelease.InvalidParameter(err).ToResp(), nil
	}

	if err := e.release.UpdateReference(orgID, releaseID, &updateReferenceRequest); err != nil {
		return apierrors.ErrUpdateRelease.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("Update succ")
}

// DeleteRelease DELETE /api/releases/<releaseId> 删除release处理
func (e *Endpoints) DeleteRelease(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getPermissionHeader(r)
	if err != nil {
		return apierrors.ErrDeleteRelease.NotLogin().ToResp(), nil
	}

	// Get releaseId
	releaseID := vars["releaseId"]
	if releaseID == "" {
		return apierrors.ErrDeleteRelease.MissingParameter("releaseId").ToResp(), nil
	}
	logrus.Infof("deleting release...releaseId: %s\n", releaseID)

	if err := e.release.Delete(orgID, releaseID); err != nil {
		return apierrors.ErrDeleteRelease.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("Delete succ")
}

// GetRelease GET /api/releases/<releaseId> release详情处理
func (e *Endpoints) GetRelease(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := getPermissionHeader(r)
	if err != nil {
		return apierrors.ErrGetRelease.NotLogin().ToResp(), nil
	}

	releaseID := vars["releaseId"]
	if releaseID == "" {
		return apierrors.ErrGetRelease.MissingParameter("releaseId").ToResp(), nil
	}
	logrus.Infof("getting release...releaseId: %s\n", releaseID)

	resp, err := e.release.Get(orgID, releaseID)
	if err != nil {
		return apierrors.ErrGetRelease.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(resp)
}

// GetDiceYAML GET /api/releases/<releaseId>/actions/get-dice 获取dice.yml内容处理
func (e *Endpoints) GetDiceYAML(w http.ResponseWriter, r *http.Request) {
	orgID, err := getPermissionHeader(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, errcode.HeaderMissing, "Header: [User-ID] or [Org-Id] is missing.")
		return
	}

	vars := mux.Vars(r)
	releaseID := vars["releaseId"]
	if releaseID == "" {
		logrus.Warn("Param [ReleaseID] is missing when get release info.")
		response.Error(w, http.StatusBadRequest, errcode.ParamMissing, "ReleaseID is missing.")
		return
	}

	logrus.Infof("getting dice.yml...releaseId: %s\n", releaseID)

	diceYAML, err := e.release.GetDiceYAML(orgID, releaseID)
	if err != nil {
		response.Error(w, http.StatusNotFound, errcode.ResourceNotFound, "release not found")
	}

	if strings.Contains(r.Header.Get("Accept"), "application/x-yaml") {
		w.Header().Set("Content-Type", "application/x-yaml; charset=utf-8")
		w.Write([]byte(diceYAML))
	} else { // 默认: application/json
		yaml, err := diceyml.New([]byte(diceYAML), false)
		if err != nil {
			logrus.Errorf("diceyml new error: %v", err)
			response.Error(w, http.StatusInternalServerError, errcode.InternalServerError, "Parse diceyml error.")
			return
		}
		diceJSON, err := yaml.JSON()
		if err != nil {
			logrus.Errorf("diceyml marshal error: %v", err)
			response.Error(w, http.StatusInternalServerError, errcode.InternalServerError, "Parse diceyml error.")
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write([]byte(diceJSON))
	}
}

// GetPlist GET /api/releases/<releaseId>/actions/get-plist 获取ios发布类型中的下载plist配置
func (e *Endpoints) GetIosPlist(ctx context.Context, writer http.ResponseWriter, r *http.Request, vars map[string]string) error {
	orgID, err := getPermissionHeader(r)
	if err != nil {
		return apierrors.ErrGetIosPlist.AccessDenied()
	}

	releaseID := vars["releaseId"]
	if releaseID == "" {
		return apierrors.ErrGetIosPlist.MissingParameter("releaseId")
	}

	plist, err := e.release.GetIosPlist(orgID, releaseID)
	if err != nil {
		return apierrors.ErrGetIosPlist.NotFound()
	}
	writer.Write([]byte(plist))
	return nil

}

// ListReleaseName 获取给定应用下的releaseName列表
func (e *Endpoints) ListReleaseName(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrListRelease.NotLogin().ToResp(), nil
	}

	// 获取applicationID
	applicationIDStr := r.URL.Query().Get("applicationId")
	if applicationIDStr == "" {
		return apierrors.ErrListRelease.MissingParameter("applicationId").ToResp(), nil
	}
	applicationID, err := strconv.ParseInt(applicationIDStr, 10, 64)
	if err != nil { // 防止SQL注入
		return apierrors.ErrListRelease.InvalidParameter("applicationId").ToResp(), nil
	}

	releaseNames, err := e.release.GetReleaseNamesByApp(orgID, applicationID)
	if err != nil {
		return apierrors.ErrListRelease.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(releaseNames)
}

// ListRelease GET /api/releases release列表处理
func (e *Endpoints) ListRelease(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListRelease.NotLogin().ToResp(), nil
	}

	params, err := e.getListParams(r, vars)
	if err != nil {
		return apierrors.ErrListRelease.InvalidParameter(err).ToResp(), nil
	}

	var orgID int64

	if !identityInfo.IsInternalClient() {
		orgID, err = getPermissionHeader(r)
		if err != nil {
			return apierrors.ErrListRelease.NotLogin().ToResp(), nil
		}

		// 获取当前用户
		userID, err := user.GetUserID(r)
		if err != nil {
			return apierrors.ErrListRelease.NotLogin().ToResp(), nil
		}

		// TODO：若没有应用的 list release 权限，则判断是否有企业权限，后续 permission list 加上 scope 后，修改该鉴权方式
		var (
			req      apistructs.PermissionCheckRequest
			permResp *apistructs.PermissionCheckResponseData
			access   bool
		)

		if params.ApplicationID > 0 {
			// 操作鉴权
			req = apistructs.PermissionCheckRequest{
				UserID:   userID.String(),
				Scope:    apistructs.AppScope,
				ScopeID:  uint64(params.ApplicationID),
				Resource: "release",
				Action:   apistructs.ListAction,
			}

			if permResp, err = e.bdl.CheckPermission(&req); err != nil {
				return apierrors.ErrListRelease.AccessDenied().ToResp(), nil
			}

			access = permResp.Access
		}

		if !access {
			req = apistructs.PermissionCheckRequest{
				UserID:   userID.String(),
				Scope:    apistructs.OrgScope,
				ScopeID:  uint64(orgID),
				Resource: "release",
				Action:   apistructs.ListAction,
			}

			if permResp, err = e.bdl.CheckPermission(&req); err != nil || !permResp.Access {
				return apierrors.ErrListRelease.AccessDenied().ToResp(), nil
			}
		}
	}

	resp, err := e.release.List(orgID, params)
	if err != nil {
		return apierrors.ErrListRelease.InternalError(err).ToResp(), nil
	}
	userIDs := make([]string, 0, len(resp.Releases))
	for _, v := range resp.Releases {
		userIDs = append(userIDs, v.UserID)
	}

	return httpserver.OkResp(resp, strutil.DedupSlice(userIDs, true))
}

// GetLatestReleases 获取指定项目指定版本情况下各应用最新release 内部使用
func (e *Endpoints) GetLatestReleases(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 检查projectID合法性
	projectIDStr := r.URL.Query().Get("projectID")
	if projectIDStr == "" {
		return apierrors.ErrListRelease.MissingParameter("projectId").ToResp(), nil
	}
	projectID, err := strutil.Atoi64(projectIDStr)
	if err != nil {
		return apierrors.ErrListRelease.InvalidParameter(err).ToResp(), nil
	}

	// 检查version合法性
	version := r.URL.Query().Get("version")
	if version == "" {
		return apierrors.ErrListRelease.MissingParameter("version").ToResp(), nil
	}

	latests, err := e.release.GetLatestReleasesByProjectAndVersion(projectID, version)
	if err != nil {
		return apierrors.ErrListRelease.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(latests)
}

// getListParams 获取 release 列表请求参数
func (e *Endpoints) getListParams(r *http.Request, vars map[string]string) (*apistructs.ReleaseListRequest, error) {
	// Get paging info
	pageSize := r.URL.Query().Get("pageSize")
	if pageSize == "" {
		pageSize = "20"
	}
	size, err := strutil.Atoi64(pageSize)
	if err != nil {
		return nil, err
	}

	pageNo := r.URL.Query().Get("pageNo")
	if pageNo == "" {
		pageNo = "1"
	}
	num, err := strutil.Atoi64(pageNo)
	if err != nil {
		return nil, err
	}

	// 按集群过滤
	clusterName := r.URL.Query().Get("cluster")

	// 按分支名过滤
	branch := r.URL.Query().Get("branchName")

	isVersion_s := r.URL.Query().Get("isVersion")
	isVersion := false
	if isVersion_s == "true" {
		isVersion = true
	}

	// 按应用过滤
	var applicationID int64
	applicationIDStr := r.URL.Query().Get("applicationId")
	if applicationIDStr != "" {
		i, err := strutil.Atoi64(applicationIDStr)
		if err != nil { // 防止SQL注入
			return nil, err
		}
		applicationID = i
	}

	// 按项目过滤
	var projectID int64
	projectIDStr := r.URL.Query().Get("projectId")
	if projectIDStr != "" {
		i, err := strutil.Atoi64(projectIDStr)
		if err != nil { // 防止SQL注入
			return nil, err
		}
		projectID = i
	}

	// filter by releaseId or releaseName
	keyword := r.URL.Query().Get("q")

	// 开始时间
	var startTime int64
	startTimeStr := r.URL.Query().Get("startTime")
	if startTimeStr != "" {
		i, err := strutil.Atoi64(startTimeStr)
		if err != nil { // 防止SQL注入
			return nil, err
		}
		startTime = i
	}
	// 结束时间
	var endTime int64
	endTimeStr := r.URL.Query().Get("endTime")
	if endTimeStr != "" {
		i, err := strutil.Atoi64(endTimeStr)
		if err != nil { // 防止SQL注入
			return nil, err
		}
		endTime = i
	} else {
		endTime = time.Now().UnixNano() / 1000 / 1000 // milliseconds
	}

	releaseName := r.URL.Query().Get("releaseName")

	crossClusterStr := r.URL.Query().Get("crossCluster")
	var crossCluster *bool
	if crossClusterStr != "" {
		b, err := strconv.ParseBool(crossClusterStr)
		if err != nil {
			return nil, err
		}
		crossCluster = &b
	}

	var crossClusterOrSpecifyCluster *string
	if s := r.URL.Query().Get("crossClusterOrSpecifyCluster"); s != "" {
		crossClusterOrSpecifyCluster = &s
	}

	return &apistructs.ReleaseListRequest{
		Query:                        keyword,
		ReleaseName:                  releaseName,
		ApplicationID:                applicationID,
		ProjectID:                    projectID,
		StartTime:                    startTime,
		EndTime:                      endTime,
		PageSize:                     size,
		PageNum:                      num,
		IsVersion:                    isVersion,
		Cluster:                      clusterName,
		Branch:                       branch,
		CrossCluster:                 crossCluster,
		CrossClusterOrSpecifyCluster: crossClusterOrSpecifyCluster,
	}, nil
}

func getPermissionHeader(r *http.Request) (int64, error) {
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		return 0, nil
	}
	return strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
}
