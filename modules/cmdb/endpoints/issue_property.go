package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver/errorresp"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/cmdb/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) CreateIssueProperty(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var createReq apistructs.IssuePropertyCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		return apierrors.ErrCreateIssuePropertyValue.InvalidParameter(err).ToResp(), nil
	}
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrCreateIssuePropertyValue.NotLogin().ToResp(), nil
	}
	createReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {

	}
	issuePropertyIndex, err := e.issueProperty.CreateProperty(&createReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(issuePropertyIndex)
}

func (e *Endpoints) DeleteIssueProperty(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var deleteReq apistructs.IssuePropertyDeleteRequest
	if err := e.queryStringDecoder.Decode(&deleteReq, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssueProperty.InvalidParameter(err).ToResp(), nil
	}
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrDeleteIssuePropertyValue.NotLogin().ToResp(), nil
	}
	deleteReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		// TODO 鉴权
	}
	property, err := e.issueProperty.GetPropertyByID(deleteReq.PropertyID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	err = e.issueProperty.DeleteProperty(property.OrgID, property.PropertyIssueType, deleteReq.PropertyID, property.Index)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(property)
}

func (e *Endpoints) UpdateIssueProperty(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var updateReq apistructs.IssuePropertyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		return apierrors.ErrDeleteIssueProperty.InvalidParameter(err).ToResp(), nil
	}
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateIssueProperty.NotLogin().ToResp(), nil
	}
	updateReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		//// check permission
		//access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
		//	UserID:   updateReq.UserID,
		//	Scope:    apistructs.AppScope,
		//	ScopeID:  uint64(updateReq.OrgID),
		//	Resource: apistructs.AppResource,
		//	Action:   apistructs.UpdateAction,
		//})
		//if err != nil {
		//	return apierrors.ErrUpdateIssueProperty.InternalError(err).ToResp(), nil
		//}
		//if !access {
		//	return apierrors.ErrUpdateIssueProperty.AccessDenied().ToResp(), nil
		//}
	}
	property, err := e.issueProperty.UpdateProperty(&updateReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(property)
}

func (e *Endpoints) UpdateIssuePropertiesIndex(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var updateReq apistructs.IssuePropertyIndexUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		return apierrors.ErrUpdateIssueProperty.InvalidParameter(err).ToResp(), nil
	}
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrUpdateIssueProperty.NotLogin().ToResp(), nil
	}
	updateReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		//// check permission
		//access, err := e.permission.CheckPermission(&apistructs.PermissionCheckRequest{
		//	UserID:   updateReq.UserID,
		//	Scope:    apistructs.AppScope,
		//	ScopeID:  uint64(updateReq.OrgID),
		//	Resource: apistructs.AppResource,
		//	Action:   apistructs.UpdateAction,
		//})
		//if err != nil {
		//	return apierrors.ErrUpdateIssueProperty.InternalError(err).ToResp(), nil
		//}
		//if !access {
		//	return apierrors.ErrUpdateIssueProperty.AccessDenied().ToResp(), nil
		//}
	}
	properties, err := e.issueProperty.UpdatePropertiesIndex(&updateReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(properties)
}

func (e *Endpoints) GetIssueProperties(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// 解析 request
	var getReq apistructs.IssuePropertiesGetRequest
	if err := e.queryStringDecoder.Decode(&getReq, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssueProperty.InvalidParameter(err).ToResp(), nil
	}
	// 鉴权
	identityInfo, err := user.GetIdentityInfo(r)
	if err != nil {
		return apierrors.ErrGetIssueProperty.NotLogin().ToResp(), nil
	}
	getReq.IdentityInfo = identityInfo
	if !identityInfo.IsInternalClient() {
		// TODO 鉴权
	}
	property, err := e.issueProperty.GetProperties(getReq)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(property)
}

func (e *Endpoints) GetIssuePropertyUpdateTime(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var req apistructs.IssuePropertyTimeGetRequest
	if err := e.queryStringDecoder.Decode(&req, r.URL.Query()); err != nil {
		return apierrors.ErrGetIssueProperty.InvalidParameter(err).ToResp(), nil
	}
	updateAt, err := e.issueProperty.GetPropertyUpdateAt(req.OrgID)
	if err != nil {
		return errorresp.ErrResp(err)
	}
	return httpserver.OkResp(updateAt)
}
