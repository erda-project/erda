package manager

import (
	"context"
	"fmt"
	"net/http"

	"github.com/erda-project/erda/modules/admin/apierrors"

	"github.com/erda-project/erda/pkg/http/httpserver"
)

func (am *AdminManager) AppendClusterEndpoint() {
	am.endpoints = append(am.endpoints, []httpserver.Endpoint{
		{Path: "/api/clusters", Method: http.MethodGet, Handler: am.ListCluster},
	}...)
}

func (am *AdminManager) ListCluster(ctx context.Context, req *http.Request, resources map[string]string) (httpserver.Responser, error) {

	var (
		orgID uint64
		err   error
	)

	orgID, err = GetOrgID(req)
	if err != nil {
		return apierrors.ErrListCluster.InvalidParameter(err).ToResp(), nil
	}

	userID := req.Header.Get("USER-ID")
	id := USERID(userID)
	if id.Invalid() {
		return apierrors.ErrListApprove.InvalidParameter(fmt.Errorf("invalid user id")).ToResp(), nil
	}

	clusterType := req.URL.Query().Get("clusterType")

	resp, err := am.bundle.ListClusters(clusterType, orgID)
	if err != nil {
		return apierrors.ErrListCluster.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp(resp)
}
