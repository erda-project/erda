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
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Endpoints) GetWorkbenchData(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var res apistructs.WorkbenchResponse
	var workReq apistructs.WorkbenchRequest
	if err := e.queryStringDecoder.Decode(&workReq, r.URL.Query()); err != nil {
		return apierrors.ErrGetWorkBenchData.InvalidParameter(err).ToResp(), nil
	}
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrGetWorkBenchData.NotLogin().ToResp(), nil
	}

	projectIDs, err := e.bdl.GetMyProjectIDs(workReq.OrgID, userID.String())
	if err != nil {
		return apierrors.ErrGetWorkBenchData.InternalError(err).ToResp(), nil
	}

	stateBelongs := []apistructs.IssueStateBelong{
		apistructs.IssueStateBelongReopen,
		apistructs.IssueStateBelongWontfix,
		apistructs.IssueStateBelongResloved,
		apistructs.IssueStateBelongWorking,
		apistructs.IssueStateBelongOpen,
	}
	stateMap := make(map[int64]dao.IssueState)
	stateReq := apistructs.IssuePagingRequest{
		OrgID:    int64(workReq.OrgID),
		PageNo:   uint64(workReq.PageNo),
		PageSize: uint64(workReq.PageSize),
		IssueListRequest: apistructs.IssueListRequest{
			StateBelongs: stateBelongs,
			Assignees:    []string{userID.String()},
			Type: []apistructs.IssueType{
				apistructs.IssueTypeRequirement,
				apistructs.IssueTypeBug,
				apistructs.IssueTypeTask,
			},
		},
	}
	if err := e.issue.FilterByStateBelongForPros(stateMap, projectIDs, &stateReq); err != nil {
		return apierrors.ErrGetWorkBenchData.InternalError(err).ToResp(), nil
	}

	resp, err := e.bdl.GetProjectListByStates(apistructs.GetProjectIDListByStatesRequest{
		StateReq: stateReq,
		ProIDs:   projectIDs,
	})
	if err != nil {
		return apierrors.ErrGetWorkBenchData.InternalError(err).ToResp(), nil
	}
	proTotal, unDonePros := resp.Total, resp.List
	res.Data.TotalProject = proTotal

	nowTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	tomorrow := nowTime.Add(time.Hour * time.Duration(24))
	twoDay := nowTime.Add(time.Hour * time.Duration(24*2))
	sevenDay := nowTime.Add(time.Hour * time.Duration(24*7))
	thirtyDay := nowTime.Add(time.Hour * time.Duration(24*30))
	timeList := [][]int64{
		{0, 0}, // not specified
		{1, nowTime.Add(time.Second * time.Duration(-1)).Unix()},                 // expired
		{nowTime.Unix(), tomorrow.Add(time.Second * time.Duration(-1)).Unix()},   // today expired
		{tomorrow.Unix(), twoDay.Add(time.Second * time.Duration(-1)).Unix()},    // tomorrow expired
		{twoDay.Unix(), sevenDay.Add(time.Second * time.Duration(-1)).Unix()},    // seven day expired
		{sevenDay.Unix(), thirtyDay.Add(time.Second * time.Duration(-1)).Unix()}, // thirty day expired
		{thirtyDay.Unix(), 0}, //feature expired
	}
	expireDays := []string{"unspecified", "expired", "oneDay", "tomorrow", "sevenDay", "thirtyDay", "feature"}

	res.Data.List = make([]apistructs.WorkbenchProjectItem, 0)
	for _, v := range unDonePros {
		var issueItem apistructs.WorkbenchProjectItem
		issueItem.IssueList = make([]apistructs.Issue, 0)
		issueReq := apistructs.IssuePagingRequest{
			OrgID:    int64(workReq.OrgID),
			PageNo:   1,
			PageSize: uint64(workReq.IssueSize),
			IssueListRequest: apistructs.IssueListRequest{
				ProjectID:    uint64(v.ID),
				StateBelongs: stateBelongs,
				Assignees:    []string{userID.String()},
				External:     true,
				OrderBy:      "plan_finished_at asc, FIELD(priority, 'URGENT', 'HIGH', 'NORMAL', 'LOW')",
				Priority: []apistructs.IssuePriority{
					apistructs.IssuePriorityUrgent,
					apistructs.IssuePriorityHigh,
					apistructs.IssuePriorityNormal,
					apistructs.IssuePriorityLow,
				},
				Type: []apistructs.IssueType{
					apistructs.IssueTypeRequirement,
					apistructs.IssueTypeBug,
					apistructs.IssueTypeTask,
				},
				Asc: true,
			},
		}
		issues, total, err := e.issue.Paging(issueReq)
		if err != nil {
			return apierrors.ErrGetWorkBenchData.InternalError(err).ToResp(), nil
		}
		issueItem.TotalIssueNum = int(total)
		issueItem.IssueList = issues
		issueItem.ProjectDTO = v

		var wg sync.WaitGroup
		wg.Add(len(expireDays))
		for index, et := range expireDays {
			go func(idx int, ed string) {
				defer wg.Done()
				etIssueReq := apistructs.IssuePagingRequest{}
				etIssueReq.OrgID = int64(workReq.OrgID)
				etIssueReq.StartFinishedAt = timeList[idx][0] * 1000
				if et == "unspecified" {
					etIssueReq.IsEmptyPlanFinishedAt = true
				} else {
					if timeList[idx][1] != 0 {
						etIssueReq.EndFinishedAt = timeList[idx][1] * 1000
					}
				}
				etIssueReq.StateBelongs = stateBelongs
				etIssueReq.ProjectID = uint64(v.ID)
				etIssueReq.PageNo = 1
				etIssueReq.PageSize = 1
				etIssueReq.External = true
				etIssueReq.Type = []apistructs.IssueType{
					apistructs.IssueTypeRequirement,
					apistructs.IssueTypeBug,
					apistructs.IssueTypeTask,
				}
				etIssueReq.Assignees = []string{userID.String()}
				_, total, err := e.issue.Paging(etIssueReq)
				if err != nil {
					logrus.Errorf("Failed to paging issue, request: %v, err: %v", etIssueReq, err)
					return
				}
				switch ed {
				case "unspecified":
					issueItem.UnSpecialIssueNum = int(total)
				case "expired":
					issueItem.ExpiredIssueNum = int(total)
				case "oneDay":
					issueItem.ExpiredOneDayNum = int(total)
				case "tomorrow":
					issueItem.ExpiredTomorrowNum = int(total)
				case "sevenDay":
					issueItem.ExpiredSevenDayNum = int(total)
				case "thirtyDay":
					issueItem.ExpiredThirtyDayNum = int(total)
				case "feature":
					issueItem.FeatureDayNum = int(total)
				}
			}(index, et)
		}
		wg.Wait()
		res.Data.List = append(res.Data.List, issueItem)
	}
	return httpserver.OkResp(res.Data)
}

func (e *Endpoints) GetIssuesForWorkbench(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var pageReq apistructs.IssuePagingRequest
	if err := e.queryStringDecoder.Decode(&pageReq, r.URL.Query()); err != nil {
		return apierrors.ErrPagingIssues.InvalidParameter(err).ToResp(), nil
	}

	switch pageReq.OrderBy {
	case "":
	case "planStartedAt":
		pageReq.OrderBy = "plan_started_at"
	case "planFinishedAt":
		pageReq.OrderBy = "plan_finished_at"
	case "assignee":
		pageReq.OrderBy = "assignee"
	default:
		return apierrors.ErrPagingIssues.InvalidParameter("orderBy").ToResp(), nil
	}

	// Authentication
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrPagingIssues.NotLogin().ToResp(), nil
	}
	pageReq.IdentityInfo = identityInfo
	pageReq.External = true

	issues, total, err := e.issue.PagingForWorkbench(pageReq)
	if err != nil {
		return apierrors.ErrPagingIssues.InternalError(err).ToResp(), nil
	}
	userIDs := pageReq.GetUserIDs()
	for _, issue := range issues {
		userIDs = append(userIDs, issue.Creator, issue.Assignee, issue.Owner)
	}
	userIDs = strutil.DedupSlice(userIDs, true)
	return httpserver.OkResp(apistructs.IssuePagingResponseData{
		Total: total,
		List:  issues,
	})
}
