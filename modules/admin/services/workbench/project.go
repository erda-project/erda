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

package workbench

import (
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	projpb "github.com/erda-project/erda-proto-go/msp/tenant/project/pb"
	"github.com/erda-project/erda/apistructs"
)

type UrlParams struct {
	Env         string `json:"env"`
	AddonId     string `json:"addonId"`
	TerminusKey string `json:"terminusKey"`
	TenantGroup string `json:"tenantGroup"`
}

func (w *Workbench) GetProjNum(identity apistructs.Identity, query string) (int, error) {
	orgID, err := strconv.Atoi(identity.OrgID)
	if err != nil {
		return 0, err
	}
	req := apistructs.ProjectListRequest{
		OrgID:    uint64(orgID),
		Query:    query,
		PageNo:   1,
		PageSize: 1,
	}
	projectDTO, err := w.bdl.ListMyProject(identity.UserID, req)
	if err != nil {
		return 0, err
	}
	if projectDTO == nil {
		return 0, nil
	}
	return projectDTO.Total, nil
}

func (w *Workbench) ListProjWbOverviewData(identity apistructs.Identity, projects []apistructs.ProjectDTO) ([]apistructs.WorkbenchProjOverviewItem, error) {
	var (
		list    []apistructs.WorkbenchProjOverviewItem
		pidList []uint64
	)
	issueMapInfo := make(map[uint64]*apistructs.WorkbenchProjectItem)
	staMapInfo := make(map[uint64]*projpb.Project)

	orgID, err := strconv.Atoi(identity.OrgID)
	if err != nil {
		return nil, err
	}

	for _, p := range projects {
		pidList = append(pidList, p.ID)
	}

	// get project issue related info
	req := apistructs.WorkbenchRequest{
		OrgID:      uint64(orgID),
		ProjectIDs: pidList,
	}
	issueInfo, err := w.bdl.GetWorkbenchData(identity.UserID, req)
	if err != nil {
		logrus.Errorf("get project workbench issue info failed, request: %+v, error: %v", req, err)
		return nil, err
	}
	for _, v := range issueInfo.Data.List {
		if v != nil {
			tmp := v
			issueMapInfo[tmp.ProjectDTO.ID] = tmp
		}
	}

	// get project msp statistic related info
	statisticInfo, err := w.bdl.GetMSPTenantProjects(identity.UserID, identity.OrgID, true, pidList)
	if err != nil {
		logrus.Errorf("get project workbench statistic info failed, request: %+v, error: %v", pidList, err)
		return nil, err
	}
	for _, v := range statisticInfo {
		if v != nil {
			tmp := v
			pid, err := strconv.Atoi(tmp.Id)
			if err != nil {
				err := fmt.Errorf("parse msp project id failed, id: %v, error: %v", tmp.Id, err)
				return nil, err
			}
			staMapInfo[uint64(pid)] = tmp
		}
	}

	// construct project workbench overview info which may contains issue/statistic info
	for i := range projects {
		p := projects[i]
		item := apistructs.WorkbenchProjOverviewItem{
			ProjectDTO: p,
		}
		is := issueMapInfo[p.ID]
		if is != nil {
			item.IssueInfo = &apistructs.ProjectIssueInfo{
				TotalIssueNum:       is.TotalIssueNum,
				UnSpecialIssueNum:   is.UnSpecialIssueNum,
				ExpiredIssueNum:     is.ExpiredIssueNum,
				ExpiredOneDayNum:    is.ExpiredOneDayNum,
				ExpiredTomorrowNum:  is.ExpiredTomorrowNum,
				ExpiredSevenDayNum:  is.ExpiredSevenDayNum,
				ExpiredThirtyDayNum: is.ExpiredThirtyDayNum,
				FeatureDayNum:       is.FeatureDayNum,
			}
		}
		st := staMapInfo[p.ID]
		if st != nil {
			item.StatisticInfo = &apistructs.ProjectStatisticInfo{
				ServiceCount:      st.ServiceCount,
				Last24HAlertCount: st.Last24HAlertCount,
			}
		}

		list = append(list, item)
	}
	return list, nil
}

func (w *Workbench) ListSubProjWbData(identity apistructs.Identity) (data *apistructs.WorkbenchProjOverviewRespData, err error) {
	var (
		projects []apistructs.ProjectDTO
		pidList  []uint64
	)
	data = &apistructs.WorkbenchProjOverviewRespData{}

	subList, err := w.bdl.ListSubscribes(identity.UserID, identity.OrgID, apistructs.GetSubscribeReq{Type: apistructs.ProjectSubscribe})
	if err != nil {
		logrus.Errorf("list subscribes failed, error: %v", err)
		return
	}
	if subList == nil || len(subList.List) == 0 {
		return
	}
	for _, v := range subList.List {
		pidList = append(pidList, v.TypeID)
	}
	rsp, err := w.bdl.GetProjectsMap(apistructs.GetModelProjectsMapRequest{ProjectIDs: pidList, KeepMsp: true})
	if err != nil {
		logrus.Errorf("get projects failed, error: %v", err)
		return nil, err
	}
	for i := range rsp {
		projects = append(projects, rsp[i])
	}

	list, err := w.ListProjWbOverviewData(identity, projects)
	if err != nil {
		logrus.Errorf("list project workbench overview data failed, error: %v", err)
		return nil, err
	}

	data = &apistructs.WorkbenchProjOverviewRespData{
		Total: len(projects),
		List:  list,
	}

	return
}

func (w *Workbench) ListQueryProjWbData(identity apistructs.Identity, page apistructs.PageRequest, query string) (data *apistructs.WorkbenchProjOverviewRespData, err error) {

	data = &apistructs.WorkbenchProjOverviewRespData{}
	orgID, err := strconv.Atoi(identity.OrgID)
	if err != nil {
		return
	}
	req := apistructs.ProjectListRequest{
		OrgID:    uint64(orgID),
		PageNo:   int(page.PageNo),
		PageSize: int(page.PageSize),
		Query:    query,
	}
	projectDTO, err := w.bdl.ListMyProject(identity.UserID, req)
	if err != nil {
		logrus.Errorf("list my project failed, request: %v, error: %v", req, err)
		return
	}
	if projectDTO == nil || len(projectDTO.List) == 0 {
		logrus.Warnf("list my project get empty response")
		return
	}

	list, err := w.ListProjWbOverviewData(identity, projectDTO.List)
	if err != nil {
		logrus.Errorf("list project workbench overview data failed, error: %v", err)
		return
	}

	data = &apistructs.WorkbenchProjOverviewRespData{
		Total: projectDTO.Total,
		List:  list,
	}

	return
}

// GetUrlCommonParams get url params used by icon
func (w *Workbench) GetUrlCommonParams(userID, orgID string, projectIDs []uint64) (urlParams []UrlParams, err error) {
	urlParams = make([]UrlParams, len(projectIDs))
	projectDTO, err := w.bdl.GetMSPTenantProjects(userID, orgID, false, projectIDs)
	if err != nil {
		logrus.Errorf("failed to get msp tenant project , err: %v", err)
		return
	}
	for i, project := range projectDTO {
		var menues []*apistructs.MenuItem
		urlParams[i].Env = project.Relationship[len(project.Relationship)-1].Workspace
		tenantId := project.Relationship[len(project.Relationship)-1].TenantID
		urlParams[i].TenantGroup = tenantId
		urlParams[i].AddonId = tenantId
		pType := project.Type
		if pType == "DOP" {
			continue
		}
		menues, err = w.bdl.ListProjectsEnvAndTenantId(userID, orgID, tenantId, pType)
		if err != nil {
			logrus.Errorf("failed to get env and tenant id ,err: %v", err)
			continue
		}
		urlParams[i].Env = project.Relationship[len(project.Relationship)-1].Workspace
		if tg, ok := menues[i].Params["tenantGroup"]; ok {
			urlParams[i].TenantGroup = tg
		}
		if tk, ok := menues[i].Params["terminusKey"]; ok {
			urlParams[i].TenantGroup = tk
		}
	}
	return
}
