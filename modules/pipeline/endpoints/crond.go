package endpoints

import (
	"context"
	"net/http"

	"github.com/erda-project/erda/pkg/httpserver"
)

func (e *Endpoints) crondReload(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	logs, err := e.crondSvc.ReloadCrond(e.pipelineSvc.RunCronPipelineFunc)
	if err != nil {
		return httpserver.ErrResp(http.StatusInternalServerError, "CROND_RELOAD", err.Error())
	}
	return httpserver.OkResp(logs)
}

func (e *Endpoints) crondSnapshot(ctx context.Context, r *http.Request, vars map[string]string) (
	httpserver.Responser, error) {

	return httpserver.OkResp(http.StatusOK, e.crondSvc.CrondSnapshot())
}
