package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
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
