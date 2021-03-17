package bundle

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/bundle/apierrors"
	"github.com/erda-project/erda/pkg/httpclient"
	"github.com/erda-project/erda/pkg/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// GetProject get project by id from cmdb.
func (b *Bundle) GetProject(id uint64) (*apistructs.ProjectDTO, error) {
	host, err := b.urls.CMDB()
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
	host, err := b.urls.CMDB()
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

// UpdateProjectActiveTime 更新项目活跃时间
func (b *Bundle) UpdateProjectActiveTime(req apistructs.ProjectActiveTimeUpdateRequest) error {
	host, err := b.urls.CMDB()
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
	host, err := b.urls.CMDB()
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
