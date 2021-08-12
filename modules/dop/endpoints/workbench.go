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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/dop/services/workbench"
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

	stateMap := make(map[int64]dao.IssueState)
	stateReq := apistructs.IssuePagingRequest{
		OrgID:    int64(workReq.OrgID),
		PageNo:   uint64(workReq.PageNo),
		PageSize: uint64(workReq.PageSize),
		IssueListRequest: apistructs.IssueListRequest{
			StateBelongs: workbench.StateBelongs,
			Assignees:    []string{userID.String()},
			Type:         workbench.IssueTypes,
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
	res.Data.List = make([]*apistructs.WorkbenchProjectItem, 0)

	for _, v := range unDonePros {
		issueItem, err := e.workBench.GetUndoneProjectItem(userID.String(), workReq.IssueSize, v)
		if err != nil {
			return apierrors.ErrGetWorkBenchData.InternalError(err).ToResp(), nil
		}
		res.Data.List = append(res.Data.List, issueItem)
	}
	if err := e.workBench.SetDiffFinishedIssueNum(stateReq, res.Data.List); err != nil {
		return apierrors.ErrGetWorkBenchData.InternalError(err).ToResp(), nil
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
