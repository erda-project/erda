package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"
)

// CreateIssue 创建事件
func (e *Endpoints) CreateIssueStage(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var createReq apistructs.IssueStageRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		return apierrors.ErrCreateIssueProperty.InvalidParameter(err).ToResp(), nil
	}
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateIssueProperty.NotLogin().ToResp(), nil
	}
	createReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		//TODO 鉴权
	}

	err = e.issue.CreateIssueStage(&createReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp("create success")
}

func (e *Endpoints) GetIssueStage(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.IssueStageRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssueBugSeverityPercentage.InvalidParameter(err).ToResp(), nil
	}
	// 需求详情
	issue, err := e.issue.GetIssueStage(&req)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(issue)
}
