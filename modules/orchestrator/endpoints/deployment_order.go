package endpoints

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/modules/orchestrator/services/apierrors"
	"github.com/erda-project/erda/pkg/http/httpserver"
	"github.com/erda-project/erda/pkg/http/httpserver/errorresp"
	"github.com/sirupsen/logrus"
	"github.com/erda-project/erda/modules/orchestrator/utils"
	"github.com/erda-project/erda/pkg/strutil"
	"strconv"
)

// CreateDeploymentOrder 创建部署请求
func (e *Endpoints) CreateDeploymentOrder(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// TODO: 需要等 pipeline action 调用走内网后，再从 header 中取 User-ID (operator)
	var (
		data *apistructs.DeploymentCreateResponseDTO
		req  apistructs.DeploymentOrderCreateRequest
	)

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// param problem
		logrus.Errorf("failed to parse request body: %v", err)
		return apierrors.ErrCreateRuntime.InvalidParameter("req body").ToResp(), nil
	}

	// TODO: auth

	if err := e.deploymentOrder.Create(&req); err != nil {
		logrus.Errorf("failed to create deployment order: %v", err)
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(data)
}

func (e *Endpoints) ListDeploymentOrder(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	// get page
	pageInfo, err := utils.GetPageInfo(r)
	if err != nil {
		return apierrors.ErrListDeploymentOrder.InvalidParameter(err).ToResp(), nil
	}

	v := r.URL.Query().Get("projectID")
	projectId, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return apierrors.ErrListDeploymentOrder.InvalidParameter(strutil.Concat("projectId: ", v)).ToResp(), nil
	}

	//orgID, err := getOrgID(r)
	//if err != nil {
	//	return apierrors.ErrListDeploymentOrder.InvalidParameter(err).ToResp(), nil
	//}
	//
	//userID, err := user.GetUserID(r)
	//if err != nil {
	//	return apierrors.ErrListDeploymentOrder.NotLogin().ToResp(), nil
	//}

	// list deployment orders
	data, err := e.deploymentOrder.List(projectId, &pageInfo)
	if err != nil {
		return errorresp.ErrResp(err)
	}

	return httpserver.OkResp(data)
}
