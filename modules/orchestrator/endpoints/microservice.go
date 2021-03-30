package endpoints

import (
	"context"
	"net/http"
	"strconv"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/modules/pkg/user"
	"github.com/erda-project/erda/pkg/httpserver"
)

// ListMicroServiceProject 获取使用微服务的项目列表
func (e *Endpoints) ListMicroServiceProject(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	_, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListAddon.NotLogin().ToResp(), nil
	}

	// 参数校验
	projectIDs := r.URL.Query()["projectId"]
	if len(projectIDs) == 0 {
		return httpserver.OkResp([]*apistructs.UniversalProjectResponseData{})
	}
	// 获取data
	data, err := e.addon.ListMicroServiceProject(projectIDs)
	if err != nil {
		return apierrors.ErrListAddon.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(data)
}

// ListMicroServiceMenu 获取项目下的微服务菜单
func (e *Endpoints) ListMicroServiceMenu(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	userID, err := user.GetUserID(r)
	if err != nil {
		return apierrors.ErrListAddon.NotLogin().ToResp(), nil
	}

	// 参数校验
	projectID := vars["projectID"]
	if projectID == "" {
		return apierrors.ErrListAddon.MissingParameter("projectID").ToResp(), nil
	}
	env := r.URL.Query().Get("env")
	if env == "" {
		return apierrors.ErrListAddon.MissingParameter("env").ToResp(), nil
	}

	projectIDInt, err := strconv.Atoi(projectID)
	if err != nil {
		return apierrors.ErrListAddon.InvalidParameter("projectID").ToResp(), nil
	}
	//鉴权
	permissionResult, err := e.bdl.CheckPermission(&apistructs.PermissionCheckRequest{
		UserID:   userID.String(),
		Scope:    apistructs.ProjectScope,
		ScopeID:  uint64(projectIDInt),
		Resource: "addon",
		Action:   apistructs.GetAction,
	})
	if err != nil {
		return apierrors.ErrListAddon.InternalError(err).ToResp(), nil
	}
	if !permissionResult.Access {
		return apierrors.ErrListAddon.AccessDenied().ToResp(), nil
	}
	// 获取data
	data, err := e.addon.ListMicroServiceMenu(projectID, env)
	if err != nil {
		return apierrors.ErrListAddon.InternalError(err).ToResp(), nil
	}

	return httpserver.OkResp(data)
}
