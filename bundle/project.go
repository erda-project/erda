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

package bundle

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/http/httpclient"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// GetProject get project by id from core-services.
func (b *Bundle) GetProject(id uint64) (*apistructs.ProjectDTO, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.ProjectDetailResponse
	resp, err := hc.Get(host, httpclient.RetryOption{}).Path(fmt.Sprintf("/api/projects/%d", id)).Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}
	if fetchResp.Data.Name == "" {
		return nil, nil
	}
	return &fetchResp.Data, nil
}

func (b *Bundle) ListProject(userID string, req apistructs.ProjectListRequest) (*apistructs.PagingProjectDTO, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp apistructs.ProjectListResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/projects")).
		Param("orgId", strconv.FormatUint(req.OrgID, 10)).
		Param("q", req.Query).
		Param("name", req.Name).
		Param("joined", strconv.FormatBool(req.Joined)).
		Param("pageNo", strconv.Itoa(req.PageNo)).
		Param("pageSize", strconv.Itoa(req.PageSize)).
		Param("isPublic", strconv.FormatBool(req.IsPublic)).
		Header("User-ID", userID).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}
	if rsp.Data.Total == 0 {
		return nil, nil
	}
	return &rsp.Data, nil
}

func (b *Bundle) ListDopProject(userID string, req apistructs.ProjectListRequest) (*apistructs.PagingProjectDTO, error) {
	host, err := b.urls.DOP()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp apistructs.ProjectListResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/projects")).
		Param("orgId", strconv.FormatUint(req.OrgID, 10)).
		Param("q", req.Query).
		Param("name", req.Name).
		Param("joined", strconv.FormatBool(req.Joined)).
		Param("pageNo", strconv.Itoa(req.PageNo)).
		Param("pageSize", strconv.Itoa(req.PageSize)).
		Param("isPublic", strconv.FormatBool(req.IsPublic)).
		Header("User-ID", userID).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}
	if rsp.Data.Total == 0 {
		return nil, nil
	}
	return &rsp.Data, nil
}

// ListMyProject 获取用户加入的项目
func (b *Bundle) ListMyProject(userID string, req apistructs.ProjectListRequest) (*apistructs.PagingProjectDTO, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp apistructs.ProjectListResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/projects/actions/list-my-projects")).
		Param("orgId", strconv.FormatUint(req.OrgID, 10)).
		Param("q", req.Query).
		Param("name", req.Name).
		Param("joined", strconv.FormatBool(req.Joined)).
		Param("pageNo", strconv.Itoa(req.PageNo)).
		Param("pageSize", strconv.Itoa(req.PageSize)).
		Param("isPublic", strconv.FormatBool(req.IsPublic)).
		Param("orderBy", req.OrderBy).
		Param("asc", strconv.FormatBool(req.Asc)).
		Header("User-ID", userID).
		Header("Org-ID", strconv.FormatUint(req.OrgID, 10)).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}
	if rsp.Data.Total == 0 {
		return nil, nil
	}
	return &rsp.Data, nil
}

// ListPublicProject 获取公开项目列表
func (b *Bundle) ListPublicProject(userID string, req apistructs.ProjectListRequest) (*apistructs.PagingProjectDTO, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp apistructs.ProjectListResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/projects/actions/list-public-projects")).
		Param("q", req.Query).
		Param("name", req.Name).
		Param("joined", strconv.FormatBool(req.Joined)).
		Param("pageNo", strconv.Itoa(req.PageNo)).
		Param("pageSize", strconv.Itoa(req.PageSize)).
		Param("isPublic", strconv.FormatBool(req.IsPublic)).
		Header("User-ID", userID).
		Header("Org-ID", strconv.FormatUint(req.OrgID, 10)).
		Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}
	if rsp.Data.Total == 0 {
		return nil, nil
	}
	return &rsp.Data, nil
}

// UpdateProjectActiveTime 更新项目活跃时间
func (b *Bundle) UpdateProjectActiveTime(req apistructs.ProjectActiveTimeUpdateRequest) error {
	host, err := b.urls.CoreServices()
	if err != nil {
		return err
	}
	hc := b.hc

	var buf bytes.Buffer
	resp, err := hc.Put(host).Path("/api/projects/actions/update-active-time").
		Header(httputil.InternalHeader, "bundle").JSONBody(&req).Do().Body(&buf)
	if err != nil {
		return apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return apierrors.ErrInvoke.InternalError(
			fmt.Errorf("failed to update project active time, status code: %d, body: %v",
				resp.StatusCode(),
				buf.String(),
			))
	}

	return nil
}

// GetWorkspaceClusterByAppBranch 根据 appID 和 branch 返回环境和集群
func (b *Bundle) GetWorkspaceClusterByAppBranch(appID uint64, gitRef string) (
	app *apistructs.ApplicationDTO,
	project *apistructs.ProjectDTO,
	branchRule *apistructs.ValidBranch,
	workspace apistructs.DiceWorkspace,
	clusterName string,
	err error,
) {
	// app
	app, err = b.GetApp(appID)
	if err != nil {
		return
	}
	// get project branch rule for workspace
	branchRule, err = b.GetBranchWorkspaceConfigByProject(app.ProjectID, gitRef)
	if err != nil {
		return
	}
	workspace = apistructs.DiceWorkspace(branchRule.Workspace)
	// get clusterName from project
	project, err = b.GetProject(app.ProjectID)
	if err != nil {
		return
	}
	for _ws, _clusterName := range project.ClusterConfig {
		if strutil.Equal(_ws, workspace.String(), true) {
			clusterName = _clusterName
			break
		}
	}
	return
}

// GetProjectNamespaceInfo 获取项目级命名空间信息
func (b *Bundle) GetProjectNamespaceInfo(projectID uint64) (*apistructs.ProjectNameSpaceInfo, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var rsp apistructs.ProjectNameSpaceInfoResponse
	resp, err := hc.Get(host).Path(fmt.Sprintf("/api/projects/%d/actions/get-ns-info", projectID)).
		Header(httputil.InternalHeader, "bundle").Do().JSON(&rsp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !rsp.Success {
		return nil, toAPIError(resp.StatusCode(), rsp.Error)
	}

	return &rsp.Data, nil
}

// IsEnabledProjectNamespaceWithoutErr 忽略错误获取项目级命名空间是否开启
func (b *Bundle) IsEnabledProjectNamespaceWithoutErr(projectID uint64) bool {
	prjNs, err := b.GetProjectNamespaceInfo(projectID)
	if err != nil {
		logrus.Errorf("[project namespace] check project namespace is enabled err: %v", err)
		return false
	}

	return prjNs.Enabled
}

// GetProjectNamespaceWithoutErr 忽略错误根据workspace获取项目级命名空间
func (b *Bundle) GetProjectNamespaceWithoutErr(projectID uint64, workspace string) string {
	var isValidWS bool
	for _, v := range apistructs.DiceWorkspaceSlice {
		if workspace == v.String() {
			isValidWS = true
			break
		}
	}

	if !isValidWS {
		return ""
	}

	prjNs, err := b.GetProjectNamespaceInfo(projectID)
	if err != nil {
		logrus.Errorf("[project namespace] get project namespace err: %v", err)
		return ""
	}

	if !prjNs.Enabled {
		return ""
	}

	return prjNs.Namespaces[workspace]
}

// GetMyProjectIDs get projectIDs by orgID adn userID from core-services.
func (b *Bundle) GetMyProjectIDs(orgID uint64, userID string) ([]uint64, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.MyProjectIDsResponse
	resp, err := hc.Get(host, httpclient.RetryOption{}).
		Path(fmt.Sprintf("/api/projects/actions/list-my-projectIDs")).
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.OrgHeader, strconv.FormatInt(int64(orgID), 10)).
		Header(httputil.UserHeader, userID).
		Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}
	return fetchResp.Data, nil
}

// GetProjectListByStates list projects by states
func (b *Bundle) GetProjectListByStates(req apistructs.GetProjectIDListByStatesRequest) (*apistructs.GetProjectIDListByStatesData, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.GetProjectIDListByStatesResponse
	resp, err := hc.Get(host).Path("/api/projects/actions/list-by-states").
		Header(httputil.InternalHeader, "bundle").JSONBody(&req).Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() || !fetchResp.Success {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return &fetchResp.Data, nil
}

// GetAllProjects get all projects
func (b *Bundle) GetAllProjects() ([]apistructs.ProjectDTO, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.GetAllProjectsResponse
	resp, err := hc.Get(host).Path("/api/projects/actions/list-all").
		Header(httputil.InternalHeader, "bundle").Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return fetchResp.Data, nil
}

// CreateProject create project
func (b *Bundle) CreateProject(req apistructs.ProjectCreateRequest, userID string) (uint64, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return 0, err
	}
	hc := b.hc

	var fetchResp apistructs.ProjectCreateResponse
	resp, err := hc.Post(host).Path("/api/projects").
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.UserHeader, userID).
		JSONBody(&req).Do().JSON(&fetchResp)
	if err != nil {
		return 0, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return 0, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return fetchResp.Data, nil
}

// DeleteProject delete project
func (b *Bundle) DeleteProject(id, orgID uint64, userID string) (*apistructs.ProjectDTO, error) {
	host, err := b.urls.CoreServices()
	if err != nil {
		return nil, err
	}
	hc := b.hc

	var fetchResp apistructs.ProjectDeleteResponse
	resp, err := hc.Delete(host).Path(fmt.Sprintf("/api/projects/%d", id)).
		Header(httputil.InternalHeader, "bundle").
		Header(httputil.UserHeader, userID).
		Header(httputil.OrgHeader, strconv.FormatUint(orgID, 10)).
		Do().JSON(&fetchResp)
	if err != nil {
		return nil, apierrors.ErrInvoke.InternalError(err)
	}
	if !resp.IsOK() {
		return nil, toAPIError(resp.StatusCode(), fetchResp.Error)
	}

	return &fetchResp.Data, nil
}
