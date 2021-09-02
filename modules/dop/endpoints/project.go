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
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/conf"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/cron"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

var (
	ProjectStatsCache *sync.Map
	Once              sync.Once
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
	if pagingProjects == nil {
		return httpserver.OkResp(&apistructs.PagingProjectDTO{})
	}

	// rich statistical data
	if params.PageSize <= 15 {
		Once.Do(func() {
			ProjectStatsCache = &sync.Map{}
		})
		for i := range pagingProjects.List {
			prjID := int64(pagingProjects.List[i].ID)
			stats, ok := ProjectStatsCache.Load(prjID)
			if !ok {
				logrus.Infof("get a new project %v add in cache", prjID)
				stats, err = e.getProjectStats(uint64(prjID))
				if err != nil {
					logrus.Errorf("fail to getProjectStats,%v", err)
					continue
				}
				ProjectStatsCache.Store(prjID, stats)
			}
			pagingProjects.List[i].Stats = *stats.(*apistructs.ProjectStats)
		}
	}

	var userIDs []string
	for _, v := range pagingProjects.List {
		userIDs = append(userIDs, v.Owners...)
	}

	return httpserver.OkResp(*pagingProjects, userIDs)
}

func (e *Endpoints) getProjectStats(projectID uint64) (*apistructs.ProjectStats, error) {
	totalApp, err := e.bdl.CountAppByProID(projectID)
	if err != nil {
		return nil, errors.Errorf("get project states err: get app err: %v", err)
	}
	totalMembers, err := e.bdl.CountMembersWithoutExtraByScope(string(apistructs.ProjectScope), projectID)
	if err != nil {
		return nil, errors.Errorf("get project states err: get member err: %v", err)
	}

	iterations, err := e.db.FindIterations(projectID)
	if err != nil {
		return nil, errors.Errorf("get project states err: get iterations err: %v", err)
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
			ProjectID: projectID,
			Type:      []apistructs.IssueType{apistructs.IssueTypeBug, apistructs.IssueTypeTask},
			External:  true,
		},
		PageNo:   1,
		PageSize: 99999,
	}, false)
	if err != nil {
		return nil, errors.Errorf("get project states err: get issues err: %v", err)
	}

	// 事件状态map
	closedBugStatsMap := make(map[int64]struct{}, 0)
	bugState, err := e.db.GetClosedBugState(int64(projectID))
	if err != nil {
		return nil, errors.Errorf("get project states err: get issues stats err: %v", err)
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
	return &apistructs.ProjectStats{
		CountApplications:       int(totalApp),
		CountMembers:            totalMembers,
		TotalApplicationsCount:  int(totalApp),
		TotalMembersCount:       totalMembers,
		TotalIterationsCount:    totalIterations,
		RunningIterationsCount:  len(runningIterations),
		PlanningIterationsCount: len(planningIterations),
		TotalManHourCount:       tManHour,
		UsedManHourCount:        uManHour,
		PlanningManHourCount:    pManHour,
		DoneBugCount:            doneBug,
		TotalBugCount:           totalBug,
		DoneBugPercent:          dBugPer,
	}, nil
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

// SetProjectStatsCache 设置项目状态缓存
func SetProjectStatsCache() {
	c := cron.New()
	if err := c.AddFunc(conf.ProjectStatsCacheCron(), func() {
		// 清空缓存
		logrus.Info("start set project stats")
		ProjectStatsCache = new(sync.Map)
	}); err != nil {
		logrus.Errorf("cron set setProjectStatsCache failed: %v", err)
	}

	c.Start()
}
