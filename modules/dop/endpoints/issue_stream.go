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
	"time"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/dop/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/strutil"
)

// CreateCommentIssueStream 创建评论活动记录
func (e *Endpoints) CreateCommentIssueStream(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	issueID, err := strutil.Atoi64(vars["id"])
	if err != nil {
		return apierrors.ErrCreateIssueStream.InvalidParameter(err).ToResp(), nil
	}

	var createReq apistructs.CommentIssueStreamCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		return apierrors.ErrCreateIssueStream.InvalidParameter(err).ToResp(), nil
	}

	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateIssueStream.NotLogin().ToResp(), nil
	}
	createReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		// TODO 鉴权
	}

	if createReq.Type == "" {
		createReq.Type = apistructs.ISTComment // 兼容现有评论
	}

	var istParam apistructs.ISTParam
	if createReq.Type == apistructs.ISTComment {
		istParam.Comment = createReq.Content
		istParam.CommentTime = time.Now().Format("2006-01-02 15:04:05")
		user, err := e.bdl.GetCurrentUser(identityInfo.UserID)
		if err != nil {
			return apierrors.ErrCreateIssueStream.InvalidParameter(err).ToResp(), nil
		}
		istParam.UserName = user.Nick
	} else { // mr 类型评论
		istParam.MRInfo = createReq.MRInfo
	}
	commentReq := &apistructs.IssueStreamCreateRequest{
		IssueID:      issueID,
		Operator:     createReq.UserID,
		StreamType:   createReq.Type,
		StreamParams: istParam,
	}
	commentID, err := e.issueStream.Create(commentReq)
	if err != nil {
		return apierrors.ErrCreateIssueStream.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(commentID)
}
