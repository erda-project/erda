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
