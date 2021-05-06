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
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

func (e *Endpoints) GetReviewByTaskId(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	var review apistructs.GetReviewByTaskIdIdRequest
	taskID, err := strutil.Atoi64(vars["taskId"])
	if err != nil {
		return apierrors.ErrGetReviewByTaskId.InvalidParameter(err).ToResp(), nil
	}
	review.TaskId = taskID
	// 创建/更新成员信息至DB
	result, err := e.ManualReview.GetReviewByTaskId(&review)
	if err != nil {
		return apierrors.ErrGetReviewByTaskId.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(apistructs.GetReviewByTaskIdIdResponse{Total: result.Total, ApprovalStatus: result.ApprovalStatus, Id: result.Id})
}

func (e *Endpoints) CreateReviewUser(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrCreateReviewUser.MissingParameter("Org-ID header is nil").ToResp(), nil
	}
	// 检查body是否为空
	if r.Body == nil {
		return apierrors.ErrCreateReviewUser.MissingParameter("body").ToResp(), nil
	}
	// 检查body格式
	var reviewCreateReq apistructs.CreateReviewUser
	if err := json.NewDecoder(r.Body).Decode(&reviewCreateReq); err != nil {
		return apierrors.ErrCreateReviewUser.InvalidParameter(err).ToResp(), nil
	}
	reviewCreateReq.OrgId = orgID
	// 创建/更新成员信息至DB

	if err := e.ManualReview.CreateReviewUser(&reviewCreateReq); err != nil {
		return apierrors.ErrCreateReviewUser.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("create review succ")
}

func (e *Endpoints) CreateReview(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrCreateReview.MissingParameter("Org-ID header is nil").ToResp(), nil
	}

	// 检查body是否为空
	if r.Body == nil {
		return apierrors.ErrCreateReview.MissingParameter("body").ToResp(), nil
	}
	// 检查body格式
	var reviewCreateReq apistructs.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&reviewCreateReq); err != nil {
		return apierrors.ErrCreateReview.InvalidParameter(err).ToResp(), nil
	}
	reviewCreateReq.OrgId = orgID
	// 创建/更新成员信息至DB
	if err := e.ManualReview.CreateOrUpdate(&reviewCreateReq); err != nil {
		return apierrors.ErrCreateReview.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("create review succ")
}

func (e *Endpoints) GetReviewsBySponsorId(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrGetReviewsBySponsorId.MissingParameter("Org-ID header is nil").ToResp(), nil
	}
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListLibReference.NotLogin().ToResp(), nil
	}
	var reviews apistructs.GetReviewsBySponsorIdRequest
	if err := e.queryStringDecoder.Decode(&reviews, r.URL.Query()); err != nil {
		return apierrors.ErrListLibReference.InvalidParameter(err).ToResp(), nil
	}
	user, _ := strconv.ParseInt(identityInfo.UserID, 10, 64)
	reviews.SponsorId = user
	reviews.OrgId = orgID
	total, list, err := e.ManualReview.GetReviewsBySponsorId(&reviews)

	uesrIDsMap := make(map[string]int)
	var userIDs []string
	for _, operators := range list {
		for _, node := range operators.Approver {
			if _, ok := uesrIDsMap[node]; ok {
				continue
			} else {
				userIDs = append(userIDs, node)
			}
		}
	}

	if err != nil {
		return apierrors.ErrGetReviewsBySponsorId.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(apistructs.ReviewsBySponsorList{List: list, Total: total}, strutil.DedupSlice(userIDs))
}

func (e *Endpoints) GetReviewsByUserId(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrGetReviewsBySponsorId.MissingParameter("Org-ID header is nil").ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrListLibReference.NotLogin().ToResp(), nil
	}

	var reviews apistructs.GetReviewsByUserIdRequest
	if err := e.queryStringDecoder.Decode(&reviews, r.URL.Query()); err != nil {
		return apierrors.ErrListLibReference.InvalidParameter(err).ToResp(), nil
	}
	user, _ := strconv.ParseInt(identityInfo.UserID, 10, 64)
	reviews.UserId = user
	reviews.OrgId = orgID
	total, list, err := e.ManualReview.GetReviewsByUserId(&reviews)
	if err != nil {
		return apierrors.ErrGetReviewsBySponsorId.InternalError(err).ToResp(), nil
	}
	var userIDs []string
	for _, node := range list {
		userIDs = append(userIDs, node.Operator)
	}
	return httpserver.OkResp(apistructs.ReviewsByUserList{List: list, Total: total}, strutil.DedupSlice(userIDs))
}

func (e *Endpoints) UpdateApproval(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrGetReviewsBySponsorId.MissingParameter("Org-ID header is nil").ToResp(), nil
	}
	if r.Body == nil {
		return apierrors.ErrUpdateApproval.MissingParameter("body").ToResp(), nil
	}
	// 检查body格式
	var reviewUpdateReq apistructs.UpdateApproval
	if err := json.NewDecoder(r.Body).Decode(&reviewUpdateReq); err != nil {
		return apierrors.ErrUpdateApproval.InvalidParameter(err).ToResp(), nil
	}
	reviewUpdateReq.OrgId = orgID
	// 创建/更新成员信息至DB
	if err := e.ManualReview.UpdateApproval(&reviewUpdateReq); err != nil {
		return apierrors.ErrUpdateApproval.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("update review succ")
}

func (e *Endpoints) GetAuthorityByUserId(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	orgID, err := strconv.ParseInt(r.Header.Get("Org-ID"), 10, 64)
	if err != nil {
		return apierrors.ErrGetReviewsBySponsorId.MissingParameter("Org-ID header is nil").ToResp(), nil
	}

	var reviews apistructs.GetAuthorityByUserIdRequest
	if err := e.queryStringDecoder.Decode(&reviews, r.URL.Query()); err != nil {
		return apierrors.ErrListLibReference.InvalidParameter(err).ToResp(), nil
	}
	reviews.OrgId = orgID
	authority, err := e.ManualReview.GetAuthorityByUserId(&reviews)
	if err != nil {
		return apierrors.ErrGetReviewsBySponsorId.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(authority)
}
