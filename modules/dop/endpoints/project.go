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
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateProject 创建项目
func (e *Endpoints) CreateProject(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 获取当前用户
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateProject.NotLogin().ToResp(), nil
	}

	if r.Body == nil {
		return apierrors.ErrCreateProject.MissingParameter("body").ToResp(), nil
	}
	var projectCreateReq apistructs.ProjectCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&projectCreateReq); err != nil {
		return apierrors.ErrCreateProject.InvalidParameter(err).ToResp(), nil
	}
	if !strutil.IsValidPrjOrAppName(projectCreateReq.Name) {
		return apierrors.ErrCreateProject.InvalidParameter(errors.Errorf("project name is invalid %s",
			projectCreateReq.Name)).ToResp(), nil
	}
	logrus.Infof("request body: %+v", projectCreateReq)

	// check permission
	req := apistructs.PermissionCheckRequest{
		UserID:   identity.UserID,
		Scope:    apistructs.OrgScope,
		ScopeID:  projectCreateReq.OrgID,
		Resource: apistructs.ProjectResource,
		Action:   apistructs.CreateAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		return apierrors.ErrCreateProject.AccessDenied().ToResp(), nil
	}

	// create project
	projectID, err := e.bdl.CreateProject(projectCreateReq, identity.UserID)
	if err != nil {
		return apierrors.ErrCreateProject.InternalError(err).ToResp(), nil
	}

	// init branchRule
	if err = e.branchRule.InitProjectRules(int64(projectID)); err != nil {
		return apierrors.ErrCreateProject.InternalError(err).ToResp(), nil
	}

	// init projectState
	if err := e.issueState.InitProjectState(int64(projectID)); err != nil {
		logrus.Warnf("failed to add state to db when create project, (%v)", err)
	}

	return httpserver.OkResp(projectID)
}

// DeleteProject delete project
func (e *Endpoints) DeleteProject(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	orgID, err := strutil.Atoi64(orgIDStr)
	if err != nil {
		return apierrors.ErrDeleteProject.InvalidParameter(err).ToResp(), nil
	}

	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteProject.NotLogin().ToResp(), nil
	}

	projectID, err := strutil.Atoi64(vars["projectID"])
	if err != nil {
		return apierrors.ErrDeleteProject.InvalidParameter(err).ToResp(), nil
	}

	// check permission
	req := apistructs.PermissionCheckRequest{
		UserID:   identity.UserID,
		Scope:    apistructs.ProjectScope,
		ScopeID:  uint64(projectID),
		Resource: apistructs.ProjectResource,
		Action:   apistructs.DeleteAction,
	}
	if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
		// check org permission
		req.Scope = apistructs.OrgScope
		req.ScopeID = uint64(orgID)
		if access, err = e.bdl.CheckPermission(&req); err != nil || !access.Access {
			return apierrors.ErrDeleteProject.AccessDenied().ToResp(), nil
		}
	}
	// delete project
	project, err := e.bdl.DeleteProject(uint64(projectID), uint64(orgID), identity.UserID)
	if err != nil {
		return apierrors.ErrDeleteProject.InternalError(err).ToResp(), nil
	}

	//  delete branch rule
	if err = e.db.DeleteBranchRuleByScope(apistructs.ProjectScope, projectID); err != nil {
		logrus.Warnf("failed to delete project branch rules, (%v)", err)
		return apierrors.ErrDeleteProject.InternalError(err).ToResp(), nil
	}

	// delete issue state
	if err = e.db.DeleteIssuesStateByProjectID(projectID); err != nil {
		logrus.Warnf("failed to delete project state, (%v)", err)
		return apierrors.ErrDeleteProject.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(project)
}

// ListProject list project
func (e *Endpoints) ListProject(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// get user
	identity, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListProject.NotLogin().ToResp(), nil
	}

	// get params
	params, err := getListProjectsParam(r)
	if err != nil {
		return apierrors.ErrListProject.InvalidParameter(err).ToResp(), nil
	}

	// check permission
	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
		req := apistructs.PermissionCheckRequest{
			UserID:   identity.UserID,
			Scope:    apistructs.OrgScope,
			ScopeID:  params.OrgID,
			Resource: apistructs.ProjectResource,
			Action:   apistructs.ListAction,
		}
		perm, err := e.bdl.CheckPermission(&req)
		if err != nil {
			return apierrors.ErrListProject.InternalError(err).ToResp(), nil
		}
		if !perm.Access {
			return apierrors.ErrListProject.AccessDenied().ToResp(), nil
		}
	}

	// list project
	pagingProjects, err := e.bdl.ListProject(identity.UserID, *params)
	if err != nil {
		return apierrors.ErrListProject.InternalError(err).ToResp(), nil
	}

	// rich statistical data
	if params.PageSize <= 15 {
		for i := range pagingProjects.List {
			if err := e.getProjectStats(pagingProjects.List[i].ID, &pagingProjects.List[i].Stats); err != nil {
				continue
			}
		}
	}

	var userIDs []string
	for _, v := range pagingProjects.List {
		userIDs = append(userIDs, v.Owners...)
	}

	return httpserver.OkResp(*pagingProjects, userIDs)
}

func (e *Endpoints) getProjectStats(projectID uint64, stat *apistructs.ProjectStats) error {
	iterations, err := e.db.FindIterations(projectID)
	if err != nil {
		return errors.Errorf("get project states err: get iterations err: %v", err)
	}
	totalIterations := len(iterations)

	runningIterations, planningIterations := make([]int64, 0), make(map[int64]bool, 0)
	now := time.Now()
	for i := 0; i < totalIterations; i++ {
		if !iterations[i].StartedAt.After(now) && iterations[i].FinishedAt.After(now) {
			runningIterations = append(runningIterations, int64(iterations[i].ID))
		}

		if iterations[i].StartedAt.After(now) {
			planningIterations[int64(iterations[i].ID)] = true
		}
	}

	var totalManHour, usedManHour, planningManHour, totalBug, doneBug int64
	totalIssues, _, err := e.db.PagingIssues(apistructs.IssuePagingRequest{
		IssueListRequest: apistructs.IssueListRequest{
			ProjectID: uint64(projectID),
			Type:      []apistructs.IssueType{apistructs.IssueTypeBug, apistructs.IssueTypeTask},
			External:  true,
		},
		PageNo:   1,
		PageSize: 99999,
	}, false)
	if err != nil {
		return errors.Errorf("get project states err: get issues err: %v", err)
	}

	// 事件状态map
	closedBugStatsMap := make(map[int64]struct{}, 0)
	bugState, err := e.db.GetClosedBugState(int64(projectID))
	if err != nil {
		return errors.Errorf("get project states err: get issues stats err: %v", err)
	}
	for _, v := range bugState {
		closedBugStatsMap[int64(v.ID)] = struct{}{}
	}

	for _, v := range totalIssues {
		var manHour apistructs.IssueManHour
		json.Unmarshal([]byte(v.ManHour), &manHour)
		// set total and used man-hour
		totalManHour += manHour.EstimateTime
		usedManHour += manHour.ElapsedTime
		// set plannning man-hour
		if _, ok := planningIterations[v.IterationID]; ok {
			planningManHour += manHour.EstimateTime
		}
		if v.Type == apistructs.IssueTypeBug {
			if _, ok := closedBugStatsMap[v.State]; ok {
				doneBug++
			}
			totalBug++
		}
	}

	tManHour, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(totalManHour)/480), 64)
	uManHour, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(usedManHour)/480), 64)
	pManHour, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(planningManHour)/480), 64)
	var dBugPer float64 = 100
	if totalBug != 0 {
		dBugPer, _ = strconv.ParseFloat(fmt.Sprintf("%.0f", float64(doneBug)*100/float64(totalBug)), 64)
	}
	stat.TotalIterationsCount = totalIterations
	stat.RunningIterationsCount = len(runningIterations)
	stat.PlanningIterationsCount = len(planningIterations)
	stat.TotalManHourCount = tManHour
	stat.UsedManHourCount = uManHour
	stat.PlanningManHourCount = pManHour
	stat.DoneBugCount = doneBug
	stat.TotalBugCount = totalBug
	stat.DoneBugPercent = dBugPer
	return nil
}

// getListProjectsParam get list project param
func getListProjectsParam(r *http.Request) (*apistructs.ProjectListRequest, error) {
	// 获取企业Id
	orgIDStr := r.Header.Get(httputil.OrgHeader)
	if orgIDStr == "" {
		orgIDStr = r.URL.Query().Get("orgId")
		if orgIDStr == "" {
			return nil, errors.Errorf("invalid param, orgId is empty")
		}
	}
	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		return nil, errors.Errorf("invalid param, orgId is invalid")
	}

	// 按项目名称搜索
	keyword := r.URL.Query().Get("q")

	// 获取pageSize
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return nil, errors.Errorf("invalid param, pageSize is invalid")
	}
	// 获取pageNo
	pageNoStr := r.URL.Query().Get("pageNo")
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	pageNo, err := strconv.Atoi(pageNoStr)
	if err != nil {
		return nil, errors.Errorf("invalid param, pageNo is invalid")
	}
	// 获取isPublic
	var isPublic bool
	isPublicStr := r.URL.Query().Get("is_public")
	if isPublicStr == "true" {
		isPublic = true
	}
	var asc bool
	ascStr := r.URL.Query().Get("asc")
	if ascStr == "true" {
		asc = true
	}
	orderBy := r.URL.Query().Get("orderBy")
	switch orderBy {
	case "cpuQuota":
		orderBy = "cpu_quota"
	case "memQuota":
		orderBy = "mem_quota"
	case "activeTime":
		orderBy = "active_time"
	case "name":
		orderBy = "name"
	default:
		orderBy = ""
	}

	return &apistructs.ProjectListRequest{
		OrgID:    uint64(orgID),
		Query:    keyword,
		Name:     r.URL.Query().Get("name"),
		PageNo:   pageNo,
		PageSize: pageSize,
		OrderBy:  orderBy,
		Asc:      asc,
		IsPublic: isPublic,
	}, nil
}
