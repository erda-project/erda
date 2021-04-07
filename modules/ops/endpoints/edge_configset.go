package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/ops/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
	"github.com/erda-project/erda/pkg/httputil"
	"github.com/erda-project/erda/pkg/strutil"
)

// ListEdgeConfigSet 获取所有边缘配置集
func (e *Endpoints) ListEdgeConfigSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var (
		pageSize    = DefaultPageSize
		pageNo      = DefaultPageNo
		orgID       int64
		clusterID   int64
		isNotPaging bool
		err         error
	)
	i, resp := e.GetIdentity(r)
	if resp != nil {
		return apierrors.ErrListEdgeApp.InternalError(fmt.Errorf("failed to get User-ID or Org-ID from request header")).ToResp(), nil
	}

	// permission check
	err = e.EdgePermissionCheck(i.UserID, i.OrgID, "", apistructs.ListAction)
	if err != nil {
		return apierrors.AccessDeny.AccessDenied().ToResp(), nil
	}

	orgIDStr := r.URL.Query().Get("orgID")
	if orgIDStr != "" {
		if orgID, err = strutil.Atoi64(orgIDStr); err != nil {
			return apierrors.ErrListEdgeConfigSet.InvalidParameter(err).ToResp(), nil
		}
	}
	clusterIDStr := r.URL.Query().Get("clusterID")
	if clusterIDStr != "" {
		if clusterID, err = strutil.Atoi64(clusterIDStr); err != nil {
			return apierrors.ErrListEdgeConfigSet.InvalidParameter(err).ToResp(), nil
		}
	}

	isNotPagingStr := r.URL.Query().Get("notPaging")
	if isNotPagingStr != "" {
		parseRes, err := strconv.ParseBool(isNotPagingStr)
		if err != nil {
			return apierrors.ErrListEdgeSite.InvalidParameter(err).ToResp(), nil
		}
		if parseRes {
			isNotPaging = true
		}
	}

	// 获取pageSize
	pageSizeStr := r.URL.Query().Get("pageSize")
	if pageSizeStr != "" {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			return apierrors.ErrListEdgeSite.InvalidParameter(err).ToResp(), nil
		}
	}

	// 获取pageNo
	pageNoStr := r.URL.Query().Get("pageNo")
	if pageNoStr != "" {
		pageNo, err = strconv.Atoi(pageNoStr)
		if err != nil {
			return apierrors.ErrListEdgeSite.InvalidParameter(err).ToResp(), nil
		}
	}

	// 参数合法性校验
	if orgID < 0 || clusterID < 0 || pageNo < 0 || pageSize < 0 {
		return apierrors.ErrListEdgeConfigSet.InternalError(fmt.Errorf("illegal query param")).ToResp(), nil
	}

	pageQueryParam := &apistructs.EdgeConfigSetListPageRequest{
		OrgID:     orgID,
		ClusterID: clusterID,
		NotPaging: isNotPaging,
		PageNo:    pageNo,
		PageSize:  pageSize,
	}

	// TODO: 操作鉴权
	total, configSetInfos, err := e.edge.ListConfigSet(pageQueryParam)

	if err != nil {
		return apierrors.ErrListEdgeConfigSet.InternalError(err).ToResp(), nil
	}

	rsp := &apistructs.EdgeConfigSetListResponse{
		Total: total,
		List:  *configSetInfos,
	}
	return httpserver.OkResp(rsp)
}

// GetEdgeConfigSet 获取指定边缘配置集
func (e *Endpoints) GetEdgeConfigSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var err error

	i, resp := e.GetIdentity(r)
	if resp != nil {
		return apierrors.ErrListEdgeApp.InternalError(fmt.Errorf("failed to get User-ID or Org-ID from request header")).ToResp(), nil
	}
	// permission check
	err = e.EdgePermissionCheck(i.UserID, i.OrgID, "", apistructs.GetAction)
	if err != nil {
		return apierrors.AccessDeny.AccessDenied().ToResp(), nil
	}

	itemID, err := strutil.Atoi64(vars["ID"])
	if err != nil {
		return apierrors.ErrListEdgeSite.InvalidParameter(err).ToResp(), nil
	}

	internalClient := r.Header.Get(httputil.InternalHeader)
	if internalClient == "" {
	}

	// TODO: 操作鉴权
	cfgSetItem, err := e.edge.GetConfigSet(itemID)

	if err != nil {
		return apierrors.ErrListEdgeSite.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(*cfgSetItem)
}

// CreateEdgeConfigSet 创建边缘配置集
func (e *Endpoints) CreateEdgeConfigSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var err error
	if r.Body == nil {
		return apierrors.ErrCreateEdgeConfigSet.MissingParameter("body").ToResp(), nil
	}
	i, resp := e.GetIdentity(r)
	if resp != nil {
		return apierrors.ErrListEdgeApp.InternalError(fmt.Errorf("failed to get User-ID or Org-ID from request header")).ToResp(), nil
	}

	// permission check
	err = e.EdgePermissionCheck(i.UserID, i.OrgID, "", apistructs.CreateAction)
	if err != nil {
		return apierrors.AccessDeny.AccessDenied().ToResp(), nil
	}
	var req apistructs.EdgeConfigSetCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrCreateEdgeConfigSet.InternalError(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", req)

	edgeSiteID, err := e.edge.CreateConfigSet(&req)
	if err != nil {
		return apierrors.ErrCreateEdgeConfigSet.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(edgeSiteID)
}

// UpdateEdgeConfigSet 更新边缘配置集
func (e *Endpoints) UpdateEdgeConfigSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var err error
	edgeSiteID, err := strutil.Atoi64(vars["ID"])
	if err != nil {
		return apierrors.ErrUpdateEdgeConfigSet.InvalidParameter(err).ToResp(), nil
	}
	i, resp := e.GetIdentity(r)
	if resp != nil {
		return apierrors.ErrListEdgeApp.InternalError(fmt.Errorf("failed to get User-ID or Org-ID from request header")).ToResp(), nil
	}

	// permission check
	err = e.EdgePermissionCheck(i.UserID, i.OrgID, "", apistructs.UpdateAction)
	if err != nil {
		return apierrors.AccessDeny.AccessDenied().ToResp(), nil
	}
	if r.Body == nil {
		return apierrors.ErrUpdateEdgeConfigSet.MissingParameter("body").ToResp(), nil
	}
	var req apistructs.EdgeConfigSetUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apierrors.ErrUpdateEdgeConfigSet.InternalError(err).ToResp(), nil
	}
	logrus.Infof("request body: %+v", req)

	err = e.edge.UpdateConfigSet(edgeSiteID, &req)
	if err != nil {
		return apierrors.ErrUpdateEdgeConfigSet.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(edgeSiteID)
}

// DeleteEdgeConfigSet 删除指定边缘配置集
func (e *Endpoints) DeleteEdgeConfigSet(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	var err error
	i, resp := e.GetIdentity(r)
	if resp != nil {
		return apierrors.ErrListEdgeApp.InternalError(fmt.Errorf("failed to get User-ID or Org-ID from request header")).ToResp(), nil
	}

	// permission check
	err = e.EdgePermissionCheck(i.UserID, i.OrgID, "", apistructs.DeleteAction)
	if err != nil {
		return apierrors.AccessDeny.AccessDenied().ToResp(), nil
	}

	configSetID, err := strutil.Atoi64(vars["ID"])
	if err != nil {
		return apierrors.ErrDeleteEdgeConfigSet.InvalidParameter(err).ToResp(), nil
	}

	if err = e.edge.DeleteConfigSet(configSetID); err != nil {
		return apierrors.ErrDeleteEdgeConfigSet.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(configSetID)
}
