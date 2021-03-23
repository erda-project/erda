package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/modules/pipeline/services/apierrors"
	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) healthCheck(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {
	_, err := e.dbClient.Exec("select 1")
	if err != nil {
		return apierrors.ErrPipelineHealthCheck.InternalError(err).ToResp(), nil
	}
	return httpserver.OkResp("success")
}
