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
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

// AddIssueRelation 添加issue关联关系
func (e *Endpoints) AddIssueRelation(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	issueID, err := strutil.Atoi64(vars["id"])
	if err != nil {
		return apierrors.ErrCreateIssueRelation.InvalidParameter(err).ToResp(), nil
	}

	var createReq apistructs.IssueRelationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		return apierrors.ErrCreateIssueRelation.InvalidParameter(err).ToResp(), nil
	}
	createReq.IssueID = uint64(issueID)
	if createReq.IssueID == createReq.RelatedIssue {
		return apierrors.ErrCreateIssueRelation.InvalidParameter("can not related yourself").ToResp(), nil
	}

	if err := createReq.Check(); err != nil {
		return apierrors.ErrCreateIssueRelation.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateIssueRelation.NotLogin().ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		// TODO 鉴权
	}

	issueRel, err := e.issueRelated.AddRelatedIssue(&createReq)
	if err != nil {
		return apierrors.ErrCreateIssueRelation.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(issueRel.ID)
}

// GetIssueRelations 获取Issue关联信息，包含关联和被关联信息
func (e *Endpoints) GetIssueRelations(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	issueID, err := strutil.Atoi64(vars["id"])
	if err != nil {
		return apierrors.ErrGetIssueRelations.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetIssueRelations.NotLogin().ToResp(), nil
	}
	if !identityInfo.IsInternalClient() {
		// TODO 鉴权
	}

	relatingIssueIDs, relatedIssueIDs, err := e.issueRelated.GetIssueRelationsByIssueIDs(uint64(issueID))
	if err != nil {
		return apierrors.ErrGetIssueRelations.InternalError(err).ToResp(), nil
	}

	relatingIssues, err := e.issue.GetIssuesByIssueIDs(relatingIssueIDs, identityInfo)
	if err != nil {
		return apierrors.ErrGetIssueRelations.InternalError(err).ToResp(), nil
	}

	relatedIssues, err := e.issue.GetIssuesByIssueIDs(relatedIssueIDs, identityInfo)
	if err != nil {
		return apierrors.ErrGetIssueRelations.InternalError(err).ToResp(), nil
	}

	// userIDs
	var userIDs []string
	for _, issue := range relatingIssues {
		userIDs = append(userIDs, issue.Creator, issue.Assignee)
	}
	for _, issue := range relatedIssues {
		userIDs = append(userIDs, issue.Creator, issue.Assignee)
	}
	userIDs = strutil.DedupSlice(userIDs, true)

	return httpserver.OkResp(apistructs.IssueRelations{
		RelatingIssues: relatingIssues,
		RelatedIssues:  relatedIssues,
	}, userIDs)
}

// DeleteIssueRelation 删除issue关联关系
func (e *Endpoints) DeleteIssueRelation(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	issueID, err := strutil.Atoi64(vars["id"])
	if err != nil {
		return apierrors.ErrDeleteIssueRelation.InvalidParameter(err).ToResp(), nil
	}
	relatedIssueID, err := strutil.Atoi64(vars["relatedIssueID"])
	if err != nil {
		return apierrors.ErrDeleteIssueRelation.InvalidParameter(err).ToResp(), nil
	}

	if err := e.issueRelated.DeleteIssueRelation(uint64(issueID), uint64(relatedIssueID)); err != nil {
		return apierrors.ErrDeleteIssueRelation.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("delete sucess")
}
