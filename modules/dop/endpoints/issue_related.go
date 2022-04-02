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

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/dao"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
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

	exist, err := e.db.IssueRelationsExist(&dao.IssueRelation{
		IssueID: createReq.IssueID,
		Type:    createReq.Type,
	}, createReq.RelatedIssue)
	if err != nil {
		return apierrors.ErrCreateIssueRelation.InternalError(err).ToResp(), nil
	}
	if exist {
		return apierrors.ErrCreateIssueRelation.AlreadyExists().ToResp(), nil
	}

	_, err = e.issueRelated.AddRelatedIssue(&createReq)
	if err != nil {
		return apierrors.ErrCreateIssueRelation.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(nil)
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

	var req apistructs.IssueRelationRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrDeleteIssueRelation.InvalidParameter(err).ToResp(), nil
	}

	if len(req.RelationTypes) == 0 {
		req.RelationTypes = []string{apistructs.IssueRelationInclusion, apistructs.IssueRelationConnection}
	}

	var userIDs []string
	var relations apistructs.IssueRelations
	for _, i := range req.RelationTypes {
		relatingIssueIDs, relatedIssueIDs, err := e.issueRelated.GetIssueRelationsByIssueIDs(uint64(issueID), []string{i})
		if err != nil {
			return apierrors.ErrGetIssueRelations.InternalError(err).ToResp(), nil
		}
		var users []string
		r := &issueRelationRetriever{
			identityInfo, relatingIssueIDs, users,
		}
		if i == apistructs.IssueRelationInclusion {
			relatingIssues, err := e.GetIssuesByRelation(r)
			if err != nil {
				return apierrors.ErrGetIssueRelations.InternalError(err).ToResp(), nil
			}
			r.issueIDs = relatedIssueIDs
			relatedIssues, err := e.GetIssuesByRelation(r)
			if err != nil {
				return apierrors.ErrGetIssueRelations.InternalError(err).ToResp(), nil
			}
			relations.IssueInclude = relatingIssues
			relations.IssueIncluded = relatedIssues
		} else {
			r.issueIDs = append(relatingIssueIDs, relatedIssueIDs...)
			relatingIssues, err := e.GetIssuesByRelation(r)
			if err != nil {
				return apierrors.ErrGetIssueRelations.InternalError(err).ToResp(), nil
			}
			relations.IssueRelate = relatingIssues
		}
		userIDs = append(userIDs, r.userIDs...)
	}
	userIDs = strutil.DedupSlice(userIDs, true)
	return httpserver.OkResp(relations, userIDs)
}

type issueRelationRetriever struct {
	identityInfo apistructs.IdentityInfo
	issueIDs     []uint64
	userIDs      []string
}

func (e *Endpoints) GetIssuesByRelation(r *issueRelationRetriever) ([]apistructs.Issue, error) {
	issues, err := e.issue.GetIssuesByIssueIDs(r.issueIDs, r.identityInfo)
	if err != nil {
		return issues, err
	}

	for _, issue := range issues {
		r.userIDs = append(r.userIDs, issue.Creator, issue.Assignee)
	}
	return issues, nil
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

	var req apistructs.IssueRelationRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrDeleteIssueRelation.InvalidParameter(err).ToResp(), nil
	}

	if err := e.issueRelated.DeleteIssueRelation(uint64(issueID), uint64(relatedIssueID), req.RelationTypes); err != nil {
		return apierrors.ErrDeleteIssueRelation.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp("delete success")
}
