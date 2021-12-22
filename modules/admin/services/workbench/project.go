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

	issueMapInfo := make(map[uint64]*apistructs.WorkbenchProjectItem)
	staMapInfo := make(map[uint64]*projpb.Project)
	list := make([]apistructs.WorkbenchProjOverviewItem, len(projects))

	orgID, err := strconv.Atoi(identity.OrgID)
	if err != nil {
		return nil, err
	}

	pidList := make([]uint64, len(projects))
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
	return list, err
}

func (w *Workbench) ListSubProjWbData(identity apistructs.Identity) (*apistructs.WorkbenchProjOverviewRespData, error) {
	var projects []apistructs.ProjectDTO
	subList, err := w.bdl.ListSubscribes(identity.UserID, identity.OrgID, apistructs.GetSubscribeReq{Type: apistructs.ProjectSubscribe})
	if err != nil {
		return nil, err
	}
	pidList := make([]uint64, len(subList.List))
	for _, v := range subList.List {
		pidList = append(pidList, v.TypeID)
	}
	rsp, err := w.bdl.GetProjectsMap(pidList)
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

	return &apistructs.WorkbenchProjOverviewRespData{
		Total: len(projects),
		List:  list,
	}, nil
}

func (w *Workbench) ListQueryProjWbData(identity apistructs.Identity, page apistructs.PageRequest, query string) (*apistructs.WorkbenchProjOverviewRespData, error) {
	orgID, err := strconv.Atoi(identity.OrgID)
	if err != nil {
		return nil, err
	}
	req := apistructs.ProjectListRequest{
		OrgID:    uint64(orgID),
		PageNo:   int(page.PageNo),
		PageSize: int(page.PageSize),
		Query:    query,
	}
	projectDTO, err := w.bdl.ListMyProject(identity.UserID, req)
	if err != nil {
		return nil, err
	}
	if projectDTO == nil {
		return nil, nil
	}

	list, err := w.ListProjWbOverviewData(identity, projectDTO.List)
	if err != nil {
		logrus.Errorf("list project workbench overview data failed, error: %v", err)
		return nil, err
	}

	return &apistructs.WorkbenchProjOverviewRespData{
		Total: projectDTO.Total,
		List:  list,
	}, nil
}
