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

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateComment 创建工单评论
func (e *Endpoints) CreateComment(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrCreateComment.InvalidParameter(err).ToResp(), nil
	}

	// 检查request body合法性
	if r.Body == nil {
		return apierrors.ErrCreateComment.MissingParameter("body is nil").ToResp(), nil
	}
	var req apistructs.CommentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateComment.InvalidParameter(err).ToResp(), nil
	}

	commentID, err := e.comment.Create(userID, &req)
	if err != nil {
		return apierrors.ErrCreateComment.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(commentID)
}

// UpdateComment 编辑工单评论
func (e *Endpoints) UpdateComment(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrUpdateComment.InvalidParameter(err).ToResp(), nil
	}

	commentIDStr := vars["commentID"]
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		return apierrors.ErrUpdateComment.InvalidParameter(err).ToResp(), nil
	}

	// 检查request body合法性
	if r.Body == nil {
		return apierrors.ErrUpdateComment.MissingParameter("body is nil").ToResp(), nil
	}
	var req apistructs.CommentUpdateRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateComment.InvalidParameter(err).ToResp(), nil
	}

	if err := e.comment.Update(commentID, userID, &req); err != nil {
		return apierrors.ErrUpdateComment.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(commentID)
}

// ListComments 工单评论列表
func (e *Endpoints) ListComments(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// 获取当前用户
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListComment.NotLogin().ToResp(), nil
	}

	ticketID, pageNo, pageSize, err := getListCommentParam(r)
	if err != nil {
		return apierrors.ErrListComment.InvalidParameter(err).ToResp(), nil
	}

	// 根据 ticketID 获取工单详情
	ticket, err := e.db.GetTicket(ticketID)
	if err != nil {
		return apierrors.ErrListComment.InternalError(err).ToResp(), nil
	}

	// 鉴权
	if ticket.TargetType == apistructs.TicketApp && userID.String() != "" {
		appID, err := strconv.ParseUint(ticket.TargetID, 10, 64)
		if err != nil {
			return apierrors.ErrListComment.InternalError(err).ToResp(), nil
		}
		req := apistructs.PermissionCheckRequest{
			UserID:   userID.String(),
			Scope:    apistructs.AppScope,
			ScopeID:  appID,
			Resource: apistructs.TicketResource,
			Action:   apistructs.OperateAction,
		}

		if access, err := e.bdl.CheckPermission(&req); err != nil || !access.Access {
			return apierrors.ErrListComment.AccessDenied().ToResp(), nil
		}
	}

	commentsResp, err := e.comment.List(ticketID, pageNo, pageSize)
	if err != nil {
		return apierrors.ErrListComment.InternalError(err).ToResp(), nil
	}
	userIDs := make([]string, 0, len(commentsResp.Comments))
	for _, v := range commentsResp.Comments {
		if v.UserID != "" {
			userIDs = append(userIDs, v.UserID)
		}
	}

	return httpserver.OkResp(*commentsResp, userIDs)
}

// DeleteComment 删除工单评论
func (e *Endpoints) DeleteComment(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	// TODO 实现
	return httpserver.OkResp(nil)
}

// 获取工单评论列表参数
func getListCommentParam(r *http.Request) (int64, int, int, error) {
	// 获取ticketID
	ticketStr := r.URL.Query().Get("ticketID")
	if ticketStr == "" {
		return 0, 0, 0, errors.Errorf("invalid param, ticket id is empty")
	}
	ticketID, err := strutil.Atoi64(ticketStr)
	if err != nil {
		return 0, 0, 0, err
	}

	// 获取pageNo参数
	pageNoStr := r.URL.Query().Get("pageNo")
	if pageNoStr == "" {
		pageNoStr = "1"
	}
	pageNo, err := strconv.Atoi(pageNoStr)
	if err != nil {
		return 0, 0, 0, errors.Errorf("invalid param, pageNo: %v", pageNoStr)
	}

	// 获取pageSize参数
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr == "" {
		pageSizeStr = "20"
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		return 0, 0, 0, errors.Errorf("invalid param, pageSize: %v", pageSizeStr)
	}

	return ticketID, pageNo, pageSize, nil
}
